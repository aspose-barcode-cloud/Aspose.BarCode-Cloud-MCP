# Component Design

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        main.go                                  │
│  ┌──────────────┐   ┌──────────────────────────────────┐        │
│  │ Env Config   │──>│ MCP Server (go-sdk)              │        │
│  │ Validation   │   │                                  │        │
│  └──────┬───────┘   │  ┌────────────────────────────┐  │        │
│         │           │  │ Tool: generate_barcode     │  │        │
│         ▼           │  ├────────────────────────────┤  │        │
│  ┌──────────────┐   │  │ Tool: recognize_barcode    │  │        │
│  │ AsposeClient │◄──│  ├────────────────────────────┤  │        │
│  │ (client.go)  │   │  │ Tool: scan_barcode         │  │        │
│  └──────────────┘   │  ├────────────────────────────┤  │        │
│         │           │  │ Tool: list_barcode_types    │  │        │
│         │           │  └────────────────────────────┘  │        │
│         ▼           └──────────────────────────────────┘        │
│  ┌──────────────┐          ▲                                    │
│  │ Aspose SDK   │          │ stdio (JSON-RPC 2.0)               │
│  │ API Client   │          ▼                                    │
│  └──────┬───────┘   ┌──────────────┐                            │
│         │           │  MCP Host    │                             │
│         ▼           └──────────────┘                            │
│  Aspose Cloud API                                               │
└─────────────────────────────────────────────────────────────────┘
```

## AsposeClient

```go
// client.go

package main

import (
    "context"
    "fmt"

    barcode "github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4"
    "golang.org/x/oauth2"
)

// AsposeClient wraps the Aspose Barcode Cloud SDK client with authentication.
type AsposeClient struct {
    API     *barcode.APIClient
    AuthCtx context.Context
}

// NewAsposeClient creates a new authenticated Aspose Barcode Cloud client.
// Returns error if credentials are empty.
func NewAsposeClient(clientID, clientSecret string) (*AsposeClient, error) {
    if clientID == "" || clientSecret == "" {
        return nil, fmt.Errorf("ASPOSE_CLIENT_ID and ASPOSE_CLIENT_SECRET must be set")
    }

    conf := barcode.NewConfiguration()
    conf.OAuthTokenUrl = "https://id.aspose.cloud/connect/token"
    conf.OAuthClientId = clientID
    conf.OAuthClientSecret = clientSecret
    apiClient := barcode.NewAPIClient(conf)

    // Create JWT-authenticated context
    // Note: Actual auth context setup depends on SDK v4 API.
    // The SDK uses oauth2 internally; the developer should verify
    // the exact pattern from SDK examples.
    jwtConf := &oauth2.Config{} // placeholder — see SDK docs
    _ = jwtConf
    authCtx := context.Background() // will be replaced with real auth context

    return &AsposeClient{
        API:     apiClient,
        AuthCtx: authCtx,
    }, nil
}
```

**Developer note**: The exact auth context setup must follow the SDK's documented pattern. See [investigation/02-aspose-barcode-cloud-api.md](../investigation/02-aspose-barcode-cloud-api.md) Authentication section for the reference code using `jwt.NewConfig()`.

## Tool Handler Pattern

Tools use the generic `mcp.AddTool` approach with typed input structs. The SDK auto-generates JSON Schema from struct tags and validates input before calling the handler:

```go
// Input struct — JSON Schema is derived from struct tags automatically
type GenerateBarcodeInput struct {
    BarcodeType string `json:"barcode_type" jsonschema:"required,description=Barcode symbology to generate"`
    Data        string `json:"data"         jsonschema:"required,description=Data to encode in the barcode"`
    ImageFormat string `json:"image_format" jsonschema:"description=Output image format,enum=PNG,enum=JPEG,enum=SVG"`
}

// Handler function — receives typed, pre-validated input
func makeGenerateHandler(client *AsposeClient) mcp.ToolHandlerFor[GenerateBarcodeInput, any] {
    return func(ctx context.Context, req *mcp.CallToolRequest, input GenerateBarcodeInput) (*mcp.CallToolResult, any, error) {
        // 1. Map input fields to SDK types
        // 2. Call Aspose SDK via client
        // 3. Return result or error
        // On error: return nil, nil, fmt.Errorf("...") — SDK wraps into IsError response
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: "result"}},
        }, nil, nil
    }
}

// Registration
mcp.AddTool(s, &mcp.Tool{
    Name:        "generate_barcode",
    Description: "Generate a barcode image...",
}, makeGenerateHandler(client))
```

Handlers are created via **factory functions** that capture the `*AsposeClient`. This avoids globals and makes testing straightforward (inject a mock client).

## Data Flow: generate_barcode

```
MCP Host                    MCP Server                          Aspose Cloud API
  │                            │                                      │
  │ tools/call                 │                                      │
  │ {generate_barcode,         │                                      │
  │  type:QR, data:"hello"}   │                                      │
  │ ──────────────────────────>│                                      │
  │                            │ Extract params                       │
  │                            │ Map to GenerateOpts                  │
  │                            │ ──────────────────────────────────── >│
  │                            │      GET /barcode/generate/QR?data=  │
  │                            │ <────────────────────────────────────│
  │                            │      200 OK (image/png bytes)        │
  │                            │ base64.StdEncoding.Encode(bytes)     │
  │ <──────────────────────────│                                      │
  │ {content: [{type:"image",  │                                      │
  │   data:"<base64>",         │                                      │
  │   mimeType:"image/png"}]}  │                                      │
```

## Data Flow: recognize_barcode / scan_barcode

```
MCP Host                    MCP Server                          Aspose Cloud API
  │                            │                                      │
  │ tools/call                 │                                      │
  │ {recognize_barcode,        │                                      │
  │  image_data:"<base64>"}   │                                      │
  │ ──────────────────────────>│                                      │
  │                            │ base64.Decode(image_data)            │
  │                            │ Write to temp file (for multipart)   │
  │                            │  OR use RecognizeBase64 endpoint     │
  │                            │ ──────────────────────────────────── >│
  │                            │      POST /barcode/recognize-body    │
  │                            │ <────────────────────────────────────│
  │                            │      200 OK (BarcodeResponseList)    │
  │                            │ Format as text                       │
  │ <──────────────────────────│                                      │
  │ {content: [{type:"text",   │                                      │
  │   text:"Found 2 barcodes: │                                      │
  │    QR: hello ..."}]}       │                                      │
```

## Error Handling

Tool errors are returned as Go `error` values from the handler. The SDK automatically wraps them into a `CallToolResult` with `IsError: true` and the error message as `TextContent`. This ensures the MCP host receives a proper error content block rather than a protocol-level error.

Required parameter validation is handled automatically by the SDK via the JSON Schema derived from input struct `jsonschema:"required"` tags.

| Error Scenario | Handling |
|---------------|----------|
| Missing required parameter | Automatic — SDK validates against JSON Schema before handler is called |
| Invalid base64 input | `return nil, nil, fmt.Errorf("invalid base64 image data: %w", err)` |
| Aspose API error (4xx/5xx) | `return nil, nil, fmt.Errorf("Aspose API error: %s", message)` |
| Network error | `return nil, nil, fmt.Errorf("failed to connect to Aspose API: %w", err)` |
| Missing credentials (startup) | `log.Fatalf()` — server refuses to start |

## Logging

- All log output goes to **stderr** (never stdout — stdout is the MCP stdio transport)
- Use `log.SetOutput(os.Stderr)` at the top of `main()`
- Log startup info: server version, credential presence (not values)
- Log each tool call at debug level (tool name, parameter summary)
