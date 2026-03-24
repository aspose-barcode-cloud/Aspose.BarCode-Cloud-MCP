# MCP (Model Context Protocol) - Technical Reference

## What is MCP

An open standard for connecting AI applications to external data sources and tools.
Think USB-C for AI: standardized protocol for connecting AI apps to external systems.

## Architecture

### Participants
- **MCP Host**: AI application (Claude Desktop, VS Code, etc.) managing one or more clients
- **MCP Client**: Component maintaining a connection to one MCP server
- **MCP Server**: Program providing context/tools to MCP clients

### Layers
1. **Data Layer**: JSON-RPC 2.0 based protocol for communication
2. **Transport Layer**: Communication mechanism (Stdio or Streamable HTTP)

## Transport Types

### Stdio (for local/Docker servers - OUR CASE)
- Uses stdin/stdout for communication
- Direct process communication, no network overhead
- Single client per server instance
- **Critical**: Never write debug output to stdout (use stderr)

### Streamable HTTP (for remote servers)
- HTTP POST for client-to-server, optional SSE for streaming
- Supports OAuth authentication

## Server Primitives (What servers expose)

### Tools (PRIMARY - what we need)
Functions that AI can invoke to perform actions.

**Tool Definition Schema:**
```json
{
  "name": "tool_name",
  "title": "Human-Readable Title",
  "description": "What the tool does",
  "inputSchema": {
    "type": "object",
    "properties": {
      "param1": {
        "type": "string",
        "description": "Parameter description"
      }
    },
    "required": ["param1"]
  }
}
```

**Tool Call Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "Result text"
    }
  ]
}
```

Content can also be `"type": "image"` with base64 data, or `"type": "resource"`.

### Resources (OPTIONAL - could be useful)
Read-only data sources (e.g., expose supported barcode types list).

### Prompts (NOT NEEDED for this project)
Reusable templates for LLM interactions.

## Protocol Lifecycle

1. **Initialize**: Client sends `initialize` request, server responds with capabilities
2. **Notification**: Client sends `notifications/initialized`
3. **Operation**: Tool discovery (`tools/list`) and execution (`tools/call`)
4. **Shutdown**: Connection termination

## Key Protocol Messages

### tools/list response
```json
{
  "tools": [
    {
      "name": "generate_barcode",
      "description": "Generate a barcode image",
      "inputSchema": { ... }
    }
  ]
}
```

### tools/call request
```json
{
  "method": "tools/call",
  "params": {
    "name": "generate_barcode",
    "arguments": {
      "type": "QR",
      "data": "Hello World"
    }
  }
}
```

### tools/call response
```json
{
  "content": [
    { "type": "image", "data": "<base64>", "mimeType": "image/png" }
  ]
}
```
