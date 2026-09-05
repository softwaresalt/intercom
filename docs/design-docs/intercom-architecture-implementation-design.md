---
status: candidate-vision
governing: false
related_decisions:
  - docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md
related_design:
  - docs/design-docs/intercom-go-backend-architecture.md
---
# **Intercom: Agent Control Plane (ACP) Architecture Design**

> **Status: candidate-vision — not governing.** This document is an
> operator-authored candidate vision input, preserved here unmodified in
> content (decision D11,
> `docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md`).
> Governance rests with design rev 2
> (`docs/design-docs/intercom-go-backend-architecture.md`) and the
> reconciliation deliberation above; this document's adoptable and deferred
> material has been carried into design rev 2's Non-Goals and "Deferred and
> adoptable material" register so it survives independently of this file.
> Note: this document's title and body use the term "ACP" (Agent Control
> Plane / Agent Client Protocol). Per decision D12 that acronym is **retired
> in both expansions** and must never appear as a Go identifier or TOML key
> in any governing artifact going forward; it is left as-is here only because
> this file's content is preserved byte-for-byte.

## **1\. Executive Summary & System Tenets**

intercom is a Go-based Man-in-the-Middle (MITM) sidecar and process supervisor designed to host the GitHub Copilot SDK natively. It replaces the ephemeral execution model of the standard copilot.exe CLI, providing strict lifecycle ownership of embedded database sidecars, unified state projection to local and remote user interfaces, and a "fail-closed" human-in-the-loop escalation protocol for autonomous agent convergence.

### **Core Architectural Tenets**

1. **Native SDK Hosting:** intercom does not wrap a CLI; it imports and executes the Copilot SDK in Go, managing raw token streams and tool execution loops.  
2. **Persistent Supervisor:** Sidecars (agent-engram, graphtor-docs, backlogit) are long-lived child processes accessed via Windows Named Pipes to eliminate database lock contention.  
3. **State Projection, Not Mutation:** UIs (TUI and Mobile PWA) are read-only projections of the central Go event bus. Commands route through the bus; UIs never mutate local state directly.  
4. **Clean-Room Execution:** Every retry, fix cycle, or phase delegation spawns a fresh, ephemeral SDK session to prevent context window degradation and token death spirals.  
5. **Fail-Closed Convergence:** The system halts and requests human authorization upon detecting divergence; it never autonomously decomposes tasks while in a degraded state.

## **2\. Process Supervision & IPC Layer**

intercom acts as the master process manager for the workspace.

### **2.1 Child Process Lifecycle**

* Upon boot, intercom uses os/exec to spawn agent-engram (Rust), graphtor-docs (Rust), and backlogit (Go) in the background.  
* **Transport:** Communication occurs exclusively over Windows Named Pipes (e.g., \\\\.\\pipe\\intercom-engram-mcp, \\\\.\\pipe\\intercom-backlogit).  
* **Resilience:** intercom implements a background health-check Goroutine. If a pipe breaks (io.EOF), intercom buffers incoming requests, silently respawns the binary, re-establishes the pipe, and flushes the buffer.

### **2.2 JSON-RPC Multiplexing**

* intercom provides a single virtual MCP/tool interface to the active Copilot SDK session.  
* When the agent yields a tool call, intercom intercepts it, routes the JSON-RPC payload to the appropriate Named Pipe, awaits the execution, and returns the response to the SDK.

## **3\. Execution Engine & Clean-Room Sessions**

The core loop abandons the "infinite context" model in favor of highly isolated execution graphs.

### **3.1 The Orchestrator Loop**

* A long-running \_orchestrator session acts as the primary planner.  
* It does *not* read code or run tests directly. It uses tools to delegate work.

### **3.2 Ephemeral Delegation (\_stage / \_ship)**

* When the orchestrator calls delegate\_to\_stage, intercom pauses the orchestrator's event channel.  
* intercom instantiates a *new*, headless Copilot SDK session loaded with \_stage.agent.md.  
* This ephemeral session accesses the Named Pipes to perform file I/O and syntax checking.  
* Upon completion, the session is destroyed (deleteSession()), and a compacted JSON summary is returned to the orchestrator.

## **4\. Convergence & Fail-Closed Escalation**

intercom continuously evaluates multi-dimensional convergence metrics during execution loops.

### **4.1 Convergence Vector Analysis**

* **Error Topology:** Tracking the shift in compiler/linter error hashes (localizing vs. spreading).  
* **Blast Radius:** Querying agent-engram via Named Pipe to measure git diff surface area against dependency graphs.  
* **Cycle Detection:** Hashing previous error states to detect oscillation loops.

### **4.2 The Fail-Closed Circuit Breaker**

If the vector indicates divergence:

1. **Halt & Rollback:** The ephemeral session halts. intercom resets the worktree to the last known stable git checkpoint.  
2. **Diagnostic Payload:** The orchestrator drafts a decomposition plan (splitting the backlogit task).  
3. **UI Handoff:** intercom locks execution and pushes the Diagnostic Payload to the TUI and Mobile PWA.  
4. **Timeout Stash:** If no human authorization is received within a configurable TTL (e.g., 15 minutes), intercom moves the item to the backlogit stash, clears the lock, and shifts the orchestrator to the next queue item.

## **5\. Telemetry & Sparse Event Logging**

Data is bifurcated to prevent disk duplication and isolate concerns.

### **5.1 Native SDK Storage (.copilot/)**

* Managed entirely by the Copilot SDK.  
* Stores raw prompt tokens, completions, and model traces within .copilot/session-state/\<session\_id\>/.

### **5.2 Intercom Telemetry (.intercom/)**

* **Session State (.intercom/session.json):** Atomic file updates containing UI session bindings and iOS VAPID push subscriptions.  
* **Sparse Event Log (.intercom/events.jsonl):** Append-only log of system markers (e.g., phase transitions, IPC restarts, UI handoffs).  
* **Correlation Schema:** Every log entry must contain: workspace\_id, task\_id (from backlogit), session\_id (from SDK), and turn\_id.

## **6\. User Interfaces (Agentic UX)**

Both interfaces are driven by the central Go event bus (Go channels syncing to WebSocket/TUI loops).

### **6.1 The Bubbletea TUI (Local)**

* **Architecture:** Elm architecture (tea.Model, tea.Update, tea.View).  
* **Layout:** 2D Command Center using lipgloss. Global health header top, nested execution cards (Main Feed \-\> Ephemeral Delegations) below.  
* **Modals:** Handoff requests dim the background and render centered, high-priority intervention modals.

### **6.2 The Mobile PWA (Remote via Dev Tunnels)**

* **Transport:** React application communicating via WebSocket through a Microsoft Dev Tunnel.  
* **Drill-Down UX:** Bottom-sheet components for reading specific task logs or phase outputs. Utilizes semantic HTML mapping from AST for iOS VoiceOver accessibility.  
* **Notifications (Web Push):** Go sidecar implements webpush-go to generate VAPID keys. The PWA registers a Service Worker on iOS to receive native lock-screen notifications for Fail-Closed handoffs.

## **7\. Implementation Sequence for AI Agent**

To build this system, execute tasks in the following strict dependency order:

### **Phase 1: IPC Supervisor & Sidecar Foundation**

1. **Refactor Rust/Go Sidecars:** Update agent-engram, graphtor-docs, and backlogit CLI entry points to support binding to Windows Named Pipes (\\\\.\\pipe\\...) in persistent listener loops.  
2. **Supervisor Bootstrapper:** Create the Go intercom package to spawn os/exec background processes and establish health-check ping loops over go-winio named pipes.  
3. **JSON-RPC Router:** Build the Go multiplexer to marshal/unmarshal JSON-RPC requests across the active pipes.

### **Phase 2: SDK Integration & Clean-Room Engine**

4. **SDK Native Host:** Import the GitHub Copilot Go SDK. Establish the initialization flow respecting COPILOT\_HOME / .copilot/.  
5. **Tool Call Interceptor:** Wire the SDK's tool execution callbacks to the JSON-RPC Router built in Phase 1\.  
6. **Clean-Room Manager:** Implement the logic to create, pause, and destroy ephemeral SDK sessions for \_stage and \_ship delegations, returning compacted JSON summaries.

### **Phase 3: Telemetry & State Bus**

7. **Sparse Event Logger:** Implement the JSONL append-only writer in .intercom/events.jsonl utilizing the task\_id \+ session\_id correlation schema.  
8. **Central Event Bus & Pub/Sub Router:** Create the core Go select loop using native channels to act as the non-blocking nervous system of the daemon.  
   * **Unidirectional Fan-Out:** Implement a generic subscriber interface that allows decoupled consumers (the Bubbletea TUI, the Dev Tunnel WebSocket broadcaster, and the JSONL logger) to register for state updates.  
   * **Strict Payload Typing:** Define rigid Go structs for all events (e.g., AgentStateChange, ToolCallRequest, HandoffIntervention) that can be seamlessly routed as tea.Msg interfaces to the TUI and marshaled into JSON for the PWA.  
   * **Concurrency Protections:** Utilize buffered channels, context.Context for graceful cancellation, and timeout thresholds to guarantee that a dropped mobile connection or slow I/O write never stalls the primary SDK execution loop. This ensures both interfaces remain perfectly synchronized, read-only projections of the true engine state.

### **Phase 4: Convergence & Handoff Circuit Breaker**

9. **Vector Analyzer:** Implement the multi-dimensional convergence logic (diff size, error hashing, oscillation detection).  
10. **Fail-Closed Escaler:** Build the workflow to halt SDK execution, perform git reset, generate the Diagnostic Payload, and initiate the UI lock/timeout-to-stash mechanism.

### **Phase 5: User Interfaces**

11. **Bubbletea TUI:** Connect a bubbletea model to the Event Bus. Build the 2D layout and intervention modals.  
12. **WebSocket & PWA Skeleton:** Expose the Event Bus over a WebSocket server. Initialize the React PWA frontend.  
13. **VAPID Web Push:** Implement the VAPID key generation and Service Worker push notification pipeline for iOS.  
14. **PWA Drill-Down UX:** Build the semantic HTML bottom-sheet UI in React, hydrated by on-demand queries to the intercom sparse log API.