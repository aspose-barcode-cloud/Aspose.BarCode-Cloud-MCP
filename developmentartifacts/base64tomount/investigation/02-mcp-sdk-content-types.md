# MCP Go SDK v0.45.0: Content Types for File Handling

## Available Content Types in CallToolResult

The `mcp.CallToolResult.Content` field accepts a `[]mcp.Content` slice. The SDK (github.com/mark3labs/mcp-go v0.45.0) supports 5 content types:

### 1. TextContent (currently used)
```go
mcp.TextContent{
    Type: "text",   // ContentTypeText
    Text: "...",
}
```

### 2. ImageContent (currently used for generated barcodes)
```go
mcp.ImageContent{
    Type:     "image",      // ContentTypeImage
    Data:     "base64...",  // base64-encoded image data
    MIMEType: "image/png",
}
```

### 3. AudioContent
```go
mcp.AudioContent{
    Type:     "audio",      // ContentTypeAudio
    Data:     "base64...",  // base64-encoded audio data
    MIMEType: "audio/wav",
}
```

### 4. ResourceLink (potential candidate for file paths)
```go
mcp.ResourceLink{
    Type:        "resource_link",  // ContentTypeLink
    URI:         "file:///path/to/barcode.png",
    Name:        "Generated barcode",
    Description: "QR code encoding 'hello'",
    MIMEType:    "image/png",
}
```

**Key properties:**
- `URI` field supports `file://` URIs
- Does NOT embed file content - just a reference
- Client must be able to access the URI to retrieve the file
- Helper: `mcp.NewResourceLink(uri, name, description, mimeType)`

### 5. EmbeddedResource (alternative for file-based content)
```go
mcp.EmbeddedResource{
    Type: "resource",  // ContentTypeResource
    Resource: mcp.TextResourceContents{
        URI:      "file:///path/to/file.svg",
        MIMEType: "image/svg+xml",
        Text:     "...",
    },
    // OR
    Resource: mcp.BlobResourceContents{
        URI:      "file:///path/to/barcode.png",
        MIMEType: "image/png",
        Blob:     "base64...",  // still base64 but with URI metadata
    },
}
```

**Key properties:**
- `TextResourceContents` - inline text content with URI metadata
- `BlobResourceContents` - base64 blob with URI metadata (still base64!)
- Helper: `mcp.NewToolResultResource(text, resourceContents)`

## Analysis for Mount-Based Approach

| Content Type | Embeds Data? | Supports file:// URI? | Relevant? |
|---|---|---|---|
| `TextContent` | Yes (text) | No | No change needed for text results |
| `ImageContent` | Yes (base64) | No | Current approach - would be replaced |
| `ResourceLink` | No (URI only) | Yes | **Best candidate** - returns path, client reads file |
| `EmbeddedResource` | Yes | Yes (metadata) | Still embeds data, adds URI as metadata |

## ResourceLink: SDK Example from examples/everything/main.go

```go
mcp.NewResourceLink(
    fmt.Sprintf("file:///example/%s.pdf", resourceType),
    fmt.Sprintf("Sample %s", resourceType),
    fmt.Sprintf("A sample %s for demonstration", resourceType),
    "application/pdf",
)
```

## Client Support Considerations

**ResourceLink** is part of the MCP specification, but client support varies:
- The MCP host must understand `resource_link` content type
- The host must have access to the `file://` URI (requires shared filesystem/mount)
- Not all MCP clients may render `resource_link` the same way as `ImageContent`
- Claude Desktop, VS Code Copilot support for `resource_link` needs verification

**ImageContent** is universally supported by all MCP clients because it's self-contained (data embedded in the message).
