---
title: "Testing a Dev Tunnel with a TUI-only credential, and the tunnel edge token"
date: 2026-10-20
tags: ["devtunnel", "tunnel", "testing", "credentials", "websocket"]
shipment: 062-S
severity: medium
---

# Dev Tunnel live-test lessons (062-S)

## 1. A private tunnel needs a second token

A client of a private Dev Tunnel must send a Dev Tunnels access token in `X-Tunnel-Authorization: tunnel <token>`, obtained with `devtunnel token <tunnel-id> --scopes connect` (about 24 hours of lifetime). The intercom credential stays in `Authorization`. Without the edge token the edge answers HTTP 200 with a sign-in page and never upgrades the WebSocket, which looks like a client bug rather than an auth refusal. Check the edge token first when a tunnelled connection returns a page instead of a 401.

## 2. Test the tunnel path without driving the TUI

The credentials appear only on the interactive TUI, so a scripted run cannot read them from the real binary. Write a temporary, uncommitted harness that wires the real agent, hub and transport, obtains credentials in-process, and sets `[remote].allowed_hosts` to the tunnel hostname. Connect with the reference client. In the 062-S run this verified the tunnel, header passthrough and a full turn with a `read` permission, but not the TUI-launched binary; record that gap rather than claiming it. Delete the harness and the tunnel afterwards, and keep tokens out of logs and the repository.
