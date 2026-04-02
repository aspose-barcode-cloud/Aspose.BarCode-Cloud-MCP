# MCP-Go SDK Reference

## Library

`github.com/mark3labs/mcp-go` - The standard Go SDK for building MCP servers.

## Installation

```bash
go get github.com/mark3labs/mcp-go
```

## Core Pattern: Server Creation

```go
import (
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

s := server.NewMCPServer(
    "Server Name",
    "1.0.0",
    server.WithToolCapabilities(false),
    server.WithRecovery(),
)

// Register tools...

// Start serving via stdio
if err := server.ServeStdio(s); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

## Defining Tools

### Tool with Parameters

```go
tool := mcp.NewTool("tool_name",
    mcp.WithDescription("What the tool does"),
    mcp.WithString("param1",
        mcp.Required(),
        mcp.Description("Parameter description"),
    ),
    mcp.WithString("param2",
        mcp.Description("Optional parameter"),
        mcp.Enum("value1", "value2", "value3"),
    ),
    mcp.WithNumber("numeric_param",
        mcp.Description("A number"),
    ),
    mcp.WithBoolean("flag_param",
        mcp.Description("A boolean"),
    ),
)
```

### Registering Tool Handler

```go
s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Extract parameters
    name, err := request.RequireString("param1")
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }

    optionalVal := request.GetString("param2", "default")

    // Do work...

    return mcp.NewToolResultText("Result text"), nil
})
```

## Request Helpers

```go
// Required (returns error if missing)
request.RequireString(name) (string, error)
request.RequireFloat(name) (float64, error)
request.RequireInt(name) (int, error)
request.RequireBool(name) (bool, error)

// Optional (returns default if missing)
request.GetString(name, defaultValue) string
request.GetFloat(name, defaultValue) float64
request.GetArguments() map[string]interface{}
request.GetStringSlice(name, defaultValue) []string
```

## Response Types

```go
mcp.NewToolResultText(text string) *mcp.CallToolResult        // text response
mcp.NewToolResultError(error string) *mcp.CallToolResult       // error response
mcp.FormatNumberResult(value float64) *mcp.CallToolResult      // number response
```

For images (base64-encoded), the result needs to contain image content type.
Check if the SDK supports `mcp.NewToolResultImage()` or use raw content construction:

```go
// Image content in MCP protocol
&mcp.CallToolResult{
    Content: []mcp.Content{
        {
            Type:     "image",
            Data:     base64EncodedString,
            MimeType: "image/png",
        },
    },
}
```

## Transport

For Docker MCP Registry (local server), use **stdio** transport:

```go
if err := server.ServeStdio(s); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

## Additional Features

- **Resources**: Expose static data via `s.AddResource()`
- **Resource Templates**: Dynamic data via `s.AddResourceTemplate()`
- **Request Hooks**: Pre/post-processing middleware
- **Session Management**: Per-session tool filtering
- **Error Recovery**: Built-in via `server.WithRecovery()`
