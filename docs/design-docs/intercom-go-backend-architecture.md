# **Architecture Design: Copilot Remote Multiplexer (Backend)**

## **1\. System Overview**

This service acts as the central state broker and local Terminal User Interface (TUI) for a headless GitHub Copilot CLI session.

**Topology (Individual Process Model):**

The system is designed to run as an isolated, dedicated process per workspace. When autoharness run is invoked, it spawns a single instance of this Go binary bound to a specific, dynamically allocated local port. This ensures 1:1 parity between a terminal tab, a repository context, and a UI state, completely isolating workspace lifecycles.

## **2\. Technology Stack**

* **Language:** Go (1.21+). Chosen for its lightweight memory footprint and robust native concurrency model (goroutines and channels). Because the architecture demands an isolated process per workspace, Go's compiled nature ensures rapid cold starts without the heavy runtime overhead of Node.js or the extensive local disk footprint of Rust. Version 1.21+ is specifically targeted to leverage the built-in slog package for structured event telemetry and the slices package for efficient state manipulations.  
* **TUI Framework:** charmbracelet/bubbletea. This framework implements the functional Elm architecture (Model, Update, View), which is essential for safely decoupling the local terminal rendering loop from the highly asynchronous network streams originating from both the ACP server and the iOS client.  
  * **Styling & Components:** We utilize lipgloss to define strict, CSS-like layout boundaries, modal overlays, and color profiles directly in the terminal. The bubbles/viewport component is critical for managing the conversational transcript; it ensures that rapid, continuous token streaming from the agent does not overwhelm standard output or permanently break the terminal's native scrollback buffer.  
* **Protocol SDK:** coder/acp-go-sdk. This package acts as the strict data contract boundary between the multiplexer and the headless Copilot CLI. The Agent Client Protocol relies heavily on JSON-RPC 2.0 union types, which Go's standard library struggles to unmarshal cleanly. This SDK provides safe, generated type matchers to reliably decode polymorphic payloads (such as differing ContentBlock types) preventing runtime panics and isolating the core logic from underlying protocol version drift.  
* **WebSockets:** gorilla/websocket. The battle-tested standard for Go HTTP upgrades. It provides the reliable Upgrader needed to accept incoming Dev Tunnel HTTP traffic and elevate it into a persistent, bidirectional binary stream. Crucially, we rely on its low-level control API to configure strict ping/pong keep-alive intervals, preventing Azure's edge infrastructure from aggressively dropping the tunnel connection during extended periods of agent "thinking" or inactivity.  
* **Tunnels:** Microsoft Dev Tunnels CLI / SDK. Operates as the zero-trust transport relay. By initiating an outbound, TLS-encrypted connection from the devbox to Azure, it completely bypasses the need to configure inbound customer firewall rules. We leverage its built-in identity gating capabilities (\--allow-anonymous false) to guarantee that only traffic bearing a validated GitHub OAuth token from your specific iOS client can interact with the local Go process.

## **3\. Multi-Workspace Orchestration (autoharness integration)**

Because we utilize an individual process per workspace, the Python-based autoharness tool acts as the process manager and port allocator.

**Startup Sequence:**

1. **Port Allocation:** autoharness run identifies two free local TCP ports (e.g., Port 8080 for ACP, Port 8081 for the Go WS Server).  
2. **Subprocess Spawn:** autoharness spawns the Go Multiplexer, passing the allocated ports and target directory as environment variables or flags.  
3. **Tunnel Registration:** autoharness adds the new WS Server port (8081) to the existing, persistent Microsoft Dev Tunnel: devtunnel port create my-persistent-tunnel \-p 8081\.  
4. **Client Discovery:** The iOS app queries the Dev Tunnel API, sees port 8081 is active, and connects to its unique, Azure-generated WebSocket URL.

## **4\. Goroutine & Channel Topology (Concurrency Model)**

Within a single workspace process, a strict channel-based concurrency model prevents blocking the local TUI while servicing remote clients and the ACP server.

1. **Main Thread (Bubble Tea):** Runs the tea.NewProgram().Run() loop. It strictly consumes messages and renders the UI.  
2. **ACP Broker Goroutine:** Maintains the TCP connection to the Copilot CLI on the allocated port. It reads JSON-RPC payloads, parses them via the acp-go-sdk, and pushes them to the Broker.Broadcast channel.  
3. **HTTP/WS Listener Goroutine:** Listens on the dynamically allocated WS port. On a valid HTTP upgrade, it spawns two goroutines per client:  
   * wsReadPump: Listens for client actions (e.g., tool approvals) over the socket.  
   * wsWritePump: Listens to a dedicated client channel to push state down the socket.  
4. **Multiplexer (Hub) Goroutine:** A central select loop that listens to Broker.Broadcast, wsReadPump actions, and Client Connect/Disconnect events.

## **5\. Core Component Implementations**

### **5.1. The State Hydrator (Memory Cache)**

The Hydrator maintains an exact replica of the conversational state to serve "Late Joiner" iOS clients who connect mid-session.

**Data Structure:**

type SessionState struct {  
    mu             sync.RWMutex  
    WorkspacePath  string  
    Messages       \[\]acp.Message          // Full chat history  
    ActiveToolCall \*acp.ToolCallRequest   // Nullable. Holds the current pending tool execution  
    ActiveContext  \[\]acp.FileContext      // Currently tracked files  
}

**Hydration Flow:**

When the Hub receives a ClientConnect event, it locks the SessionState (mu.RLock()), marshals the entire struct into a JSON payload of type session/hydrate, and pushes it directly to that specific client's wsWritePump channel *before* adding the client to the live broadcast pool.

### **5.2. The Multiplexer (The Hub)**

The Multiplexer is the central router. It receives strongly-typed acp.Event interfaces from the ACP Broker.

**Message Routing:**

func (h \*Hub) Run() {  
    for {  
        select {  
        case client := \<-h.register:  
            // 1\. Send hydration state to client  
            // 2\. Add to active clients map  
        case event := \<-h.acpBroadcast:  
            // 1\. Update State Hydrator  
            // 2\. Send to Bubble Tea: program.Send(event)  
            // 3\. Loop through active WS clients and send event  
        case action := \<-h.clientActions:  
            // Handle tool approvals from iOS or local TUI  
        }  
    }  
}

**Tool Execution Lock (First-Responder):**

When a tool/request arrives, SessionState.ActiveToolCall is populated. The prompt is rendered both locally and remotely. If an approval action arrives via h.clientActions:

1. Verify the action ID matches SessionState.ActiveToolCall.  
2. If yes, forward the JSON-RPC response to the ACP Broker.  
3. Set SessionState.ActiveToolCall to nil.  
4. Broadcast a tool/resolved event to clear the UI modal on all other connected clients.

### **5.3. The Bubble Tea TUI**

The local TUI operates entirely on tea.Msg events, ensuring thread-safe rendering.

**Model State:**

The Bubble Tea Model maintains a viewport.Model for scrolling history, a textinput.Model for local prompts, and tracks the pending tool state.

**Update Loop:**

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {  
    switch msg := msg.(type) {  
    case acp.AgentStdoutEvent:  
        // Append tokens to the current conversational bubble  
        // Update viewport content  
        return m, viewport.ViewDown()  
    case acp.ToolRequestEvent:  
        // Trigger a lipgloss-styled modal overlay for Y/N approval  
        m.pendingTool \= msg  
        return m, nil  
    case ToolResolvedEvent:  
        // Another client answered (or local answered). Clear the modal.  
        m.pendingTool \= nil  
        return m, nil  
    case tea.KeyMsg:  
        // Handle local user input, package as ACP request, send to Hub  
    }  
}

