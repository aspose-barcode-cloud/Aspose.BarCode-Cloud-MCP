# Architecture Overview

## System Context

The **Aspose Barcode MCP Server** is a Go application that bridges MCP-compatible AI hosts (Claude Desktop, VS Code, etc.) with the Aspose Barcode Cloud API. It runs as a Docker container, communicates via **stdio** transport, and is distributed through the **Docker MCP Registry**.

```
┌─────────────────┐     stdio (JSON-RPC 2.0)     ┌──────────────────────┐     HTTPS/REST      ┌──────────────────────┐
│                 │ ──────────────────────────────>│                      │ ──────────────────> │                      │
│   MCP Host      │                               │  Aspose Barcode      │                     │  Aspose Barcode      │
│  (Claude, etc.) │ <──────────────────────────────│  MCP Server (Go)     │ <────────────────── │  Cloud API v4.0      │
│                 │     stdio (JSON-RPC 2.0)       │  (Docker container)  │     HTTPS/REST      │                      │
└─────────────────┘                               └──────────────────────┘                     └──────────────────────┘
                                                         │                                            │
                                                         │ reads env vars                             │ OAuth 2.0 JWT
                                                         ▼                                            ▼
                                                  ASPOSE_CLIENT_ID                         https://id.aspose.cloud
                                                  ASPOSE_CLIENT_SECRET                     /connect/token
```

## Key Design Decisions (from investigation)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Tool granularity | 4 focused tools | LLMs work better with fewer, well-described tools |
| Image transfer | Base64 strings | Universal across MCP clients, no filesystem dependency |
| Barcode type validation | Pass-through to API + discovery tool | Simpler, always up-to-date |
| Transport | stdio | Required for Docker MCP Registry local servers |
| Dockerfile | Multi-stage build | Minimal image, no Go toolchain in production |
| Configuration | Environment variables only | Docker secrets mechanism handles credential injection |

## Technology Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.24+ |
| MCP SDK | github.com/mark3labs/mcp-go | latest |
| Aspose SDK | github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4 | v4.2603.0 |
| Container | Docker multi-stage (golang:1.24-alpine → alpine) | - |
| License | MIT | - |

## MCP Tools Exposed

| Tool | Purpose | API Group |
|------|---------|-----------|
| `generate_barcode` | Generate barcode image from data | Generate API |
| `recognize_barcode` | Recognize specific barcode types from image | Recognize API |
| `scan_barcode` | Auto-detect all barcodes in image | Scan API |
| `list_barcode_types` | List supported barcode symbologies | Local (no API call) |
