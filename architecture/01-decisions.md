# Architecture Decision Records: Mount-Based File Distribution

## ADR-1: MCP Content Type for File Path Return

**Decision: Dual-mode approach (conditional on mount configuration)**

- **Mount configured** (`ASPOSE_CLOUD_MOUNT_PATH` is set): Write file to mount directory, return `mcp.TextContent` with the relative file path and metadata.
- **Mount not configured**: Check on startup and return error response with clarification.

**Rationale:**
- `TextContent` is universally supported by all MCP clients. Every host can parse a text response.
- `ResourceLink` (`file://` URI) is semantically ideal but client support is unverified for Claude Desktop, VS Code Copilot, and Docker MCP Toolkit. Adopting it risks broken rendering.
- `EmbeddedResource` with `BlobResourceContents` still embeds base64 data, defeating the purpose.

**Response format when mount is active (generate_barcode):**
```
Generated barcode image saved to: qr-20260326-143022-a1b2.png
Format: image/png
Mount path: /mnt/data/qr-20260326-143022-a1b2.png
```

**Rejected alternatives:**
- A) `ResourceLink` only -- client support risk too high
- C) `EmbeddedResource` -- still uses base64, no benefit
- D) Both ImageContent + path -- doubles payload size when mount is available

---

## ADR-2: Mount Path Configuration

**Decision: Environment variable `ASPOSE_CLOUD_MOUNT_PATH` + Docker Registry parameter**

- Server reads `ASPOSE_CLOUD_MOUNT_PATH` environment variable at startup.
- If empty/unset, server returns error response on startup.
- Docker MCP Registry `server.yaml` exposes a `data_dir` parameter that maps to a volume mount targeting this path.

**Rationale:**
- Consistent with existing credential pattern (`ASPOSE_CLOUD_CLIENT_ID`, `ASPOSE_CLOUD_CLIENT_SECRET`).
- Registry parameter integrates with Docker MCP Toolkit UI, allowing users to configure the mount path visually.
- Environment variable works for all deployment modes (Docker, binary, dev).

**Container-side default path:** `/mnt/data`

**Rejected alternatives:**
- A) Hardcoded path -- inflexible, can't disable mount mode
- C) Command-line argument -- requires ENTRYPOINT changes, harder to configure

---

## ADR-3: File Naming Strategy

**Decision: Combination -- `{type_prefix}-{timestamp}-{short_uuid}.{ext}`**

**Examples:**
- `qr-20260326-143022-a1b2c3d4.png`
- `code128-20260326-143023-e5f6g7h8.jpeg`
- `datamatrix-20260326-143024-i9j0k1l2.svg`

**Rules:**
- `type_prefix`: Lowercase barcode type name, truncated to 20 chars
- `timestamp`: `YYYYMMDD-HHMMSS` in UTC
- `short_uuid`: First 8 characters of a UUID v4
- `ext`: File extension matching the image format

**File cleanup policy:** Server does NOT clean up generated files. The host/user manages the mounted directory. This is consistent with how the filesystem MCP server works -- it provides access but doesn't manage lifecycle.

**Rationale:**
- Human-readable: operator can identify barcode type and creation time at a glance.
- Sortable: timestamp prefix enables chronological sorting.
- Collision-resistant: UUID suffix prevents conflicts even with concurrent requests.
- No user-provided filename parameter to avoid path traversal attack surface.

**Rejected alternatives:**
- A) UUID only -- not human-readable
- B) Descriptive names from data -- collision risk, data may contain special characters
- D) User-provided filename -- path traversal risk, adds API surface

---

## ADR-4: Input Parameter Change for Recognize/Scan

**Decision:  Replace `image_data` with `image_path` (string, relative to mount root)**

**New input schema:**
- `image_path` (string, optional): Relative file path within the mount directory

- Server resolves the path relative to `ASPOSE_CLOUD_MOUNT_PATH`
- Server validates the resolved path is within the mount directory (security)
- Server opens the file and uses `RecognizeMultipart` / `ScanMultipart` SDK methods

**Rationale:**
- Server is not deployed yet and breaking changed don't affect users. 
- `RecognizeMultipart` requires `barcodeType` as a separate parameter, which `recognize_barcode` already accepts.

**Rejected alternatives:**
- B) Add `image_path`** alongside `image_data`, support both - backwards compatible but complex
- C) Absolute container path -- exposes container internals
- D) Auto-detection -- implicit behavior is error-prone

---

## ADR-5: Bidirectional File Sharing

**Decision: Use mount for both directions (generation output + recognition/scan input)**

- **Generation (server -> host):** Server writes barcode image to mount, returns relative path.
- **Recognition/Scan (host -> server):** Host places image in mount, passes relative path as `image_path`.
- Both directions share the same mount point and same `ASPOSE_CLOUD_MOUNT_PATH`.

**Rationale:**
- Symmetric design: one mount, both directions.
- Eliminates base64 from the entire pipeline.
- For recognition, the AI host already has filesystem access (it's running the Docker container). Placing a file in the mount is natural.
- Single mount point simplifies Docker configuration.

---

## ADR-6: Docker MCP Registry Configuration

**Decision:**

```yaml
run:
  volumes:
    - "{{aspose-barcode-cloud.data_dir}}:/mnt/data"
  env:
    ASPOSE_CLOUD_MOUNT_PATH: "/mnt/data"
config:
  parameters:
    type: object
    properties:
      data_dir:
        type: string
        description: "Host directory for barcode file exchange"
    required:
      - data_dir
```

- Mount path is user-configurable via `data_dir` registry parameter.
- `disableNetwork: false` remains (server needs Aspose Cloud API access over HTTPS).
- No default host-side path -- user must explicitly configure it for security (same model as filesystem MCP server).
- Container-side path is fixed at `/mnt/data`.

---

## ADR-7: Security Measures

**Decision: Defense-in-depth path validation**

1. **Path traversal prevention:** All `image_path` inputs are validated:
   - `filepath.Clean()` the input path
   - Join with mount root: `filepath.Join(mountPath, cleanedPath)`
   - Verify result starts with `mountPath` prefix after cleaning
   - Reject paths containing `..` after cleaning

2. **No symlink following:** Use `filepath.EvalSymlinks()` on the resolved path and verify it's still within the mount directory.

3. **Single mount point:** One directory for both input and output. No separate read-only/write-only mounts -- this adds configuration complexity without meaningful security benefit in the MCP context (the AI host already controls both directions).

4. **File type validation for input:** Only accept files with recognized image extensions (`.png`, `.jpg`, `.jpeg`, `.gif`, `.tiff`, `.bmp`).

5. **Generated file permissions:** Write files with `0644` permissions (owner read/write, group/other read).

---

## ADR-8: Client Compatibility Strategy

**Decision: No special client-specific handling needed**

- Since we use `TextContent` (universally supported) for file path responses, all MCP clients will receive and display the result correctly.
- The AI host (Claude, Copilot, etc.) reads the text response containing the file path and can access the file through its own filesystem access.


**Client configuration examples will be documented in README for:**
- Docker MCP Toolkit (automatic via registry)
- Claude Desktop (`claude_desktop_config.json` with `--mount` args)
- VS Code (`.vscode/mcp.json` with `-v` args and `${workspaceFolder}`)
- Docker MCP Gateway (`gordon-mcp.yml` with volumes)

---

## ADR-9: Path Format in Responses

**Decision: Return relative paths (relative to mount root)**

- Server returns only the filename (or relative sub-path) of the generated file, NOT the absolute container path.
- Example: `qr-20260326-143022-a1b2c3d4.png`, NOT `/mnt/data/qr-20260326-143022-a1b2c3d4.png`

**Rationale:**
- The host knows its own mount source path. It doesn't need to know the container's internal path.
- The AI host maps the relative path to its own filesystem: `{host_mount_path}/{relative_path}`.
- Avoids leaking container internals.
- Consistent with how `image_path` input works (also relative to mount root).

---

## ADR-10: SVG Handling

**Decision: SVG writes as files like in binary formats**

- Difference in file sharing may confuse weak models.




