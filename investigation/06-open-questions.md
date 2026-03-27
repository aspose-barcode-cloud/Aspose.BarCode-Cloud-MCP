# Open Questions for System Architect

## Q1: MCP Content Type for File Path Return

**Context:** The MCP Go SDK v0.45.0 offers several content types. When `generate_barcode` returns a file path instead of base64 data, which content type should be used?

**Options:**
- **A) `mcp.ResourceLink`** with `file://` URI - pure reference, no embedded data. Client must access the file at that URI.
- **B) `mcp.TextContent`** with the file path as plain text - simplest, universally supported, but no MIME type metadata.
- **C) `mcp.EmbeddedResource`** with `BlobResourceContents` - still embeds base64 data but adds URI metadata (defeats the purpose of mount-based approach).
- **D) Keep `mcp.ImageContent`** with base64 for backwards compatibility AND additionally return the file path via `mcp.TextContent` or `mcp.ResourceLink`.

**Tradeoff:** `ResourceLink` is semantically correct but client support may vary. `TextContent` is universally supported but loses type information.

## Q2: Mount Path Configuration

**Context:** The server needs to know where the mounted directory is inside the container.

**Options:**
- **A) Hardcoded path** (e.g., `/mnt/data` or `/workspace`) - simple but inflexible
- **B) Environment variable** (e.g., `ASPOSE_CLOUD_MOUNT_PATH`) - consistent with existing credential pattern
- **C) Command-line argument** - requires changes to ENTRYPOINT/CMD
- **D) Registry parameter** via `config.parameters` in server.yaml - integrates with Docker MCP Toolkit UI

## Q3: File Naming Strategy for Generated Barcodes

**Context:** When `generate_barcode` writes a file to the mount, it needs a filename.

**Options:**
- **A) UUID-based** (e.g., `a1b2c3d4.png`) - no collisions, no meaningful names
- **B) Descriptive** (e.g., `qr-hello-world.png`) - human-readable but may collide
- **C) Timestamp-based** (e.g., `barcode-20260326-143022.png`) - sortable, low collision
- **D) User-provided** via new optional `output_filename` parameter - flexible but adds API surface
- **E) Combination** (e.g., `qr-a1b2c3d4.png`) - type prefix + UUID suffix

**Related:** Should generated files be cleaned up? By the server? By the client? Never?

## Q4: Input Parameter Change for Recognize/Scan

**Context:** Currently `recognize_barcode` and `scan_barcode` accept `image_data` as a base64 string. With mounts, they should accept a file path instead.

**Options:**
- **A) Replace `image_data`** with `image_path` (string, relative to mount root) - breaking change
- **B) Add `image_path`** alongside `image_data`, support both - backwards compatible but complex
- **C) Replace `image_data`** with `image_path` (absolute path inside container) - simpler but exposes container internals
- **D) Accept both** with auto-detection (if starts with `/` treat as path, otherwise treat as base64) - implicit, may be error-prone

## Q5: Bidirectional File Sharing

**Context:** The current proposal has two directions:
1. **Generation (server -> host):** Server writes barcode to mounted dir, returns path
2. **Recognition (host -> server):** Host places image in mounted dir, passes path to server

**Question:** Should both directions use the mount, or only generation? Recognition/scan could continue using base64 for input if the primary goal is avoiding large base64 strings in output.

## Q6: Docker MCP Registry Compatibility

**Context:** The `registry/server.yaml` needs a `run.volumes` section. The template syntax allows user-configurable paths.

**Questions:**
- Should the mount path be user-configurable via registry parameters (like the filesystem server)?
- Should `disableNetwork: false` remain (server needs Aspose API access)?
- What should the default host-side path be?

## Q7: Security Considerations

**Questions:**
- Should the server validate that file paths are within the mount directory (path traversal prevention)?
- Should file permissions be restricted (read-only for input, write-only for output)?
- Should there be separate mount points for input and output?
- What about symlink following?

## Q8: Client Compatibility Impact

**Context:** Different MCP hosts handle Docker differently.

**Questions:**
- Does Claude Desktop properly handle `ResourceLink` content type with `file://` URIs?
- Does VS Code Copilot render `ResourceLink` as clickable/viewable files?
- Will the Docker MCP Toolkit auto-configure mounts from registry metadata?
- Should the server support a "legacy mode" (base64) for clients that can't use mounts?

## Q9: Path Format Consistency

**Context:** Docker containers run Linux. Hosts may be macOS, Windows, or Linux.

**Questions:**
- Returned paths are container-internal (`/mnt/data/barcode.png`). Should the server return the host-mapped path instead?
- How does the MCP host resolve `file:///mnt/data/barcode.png` back to the host filesystem?
- Is this resolution automatic with bind mounts, or does the client need mapping logic?

## Q10: SVG Handling

**Context:** SVG is currently returned as `mcp.TextContent` (raw SVG string), not base64. It's text, not binary.

**Question:** Should SVG also be written to a file and returned as a path? Or should it remain inline as text since it's already efficient?
