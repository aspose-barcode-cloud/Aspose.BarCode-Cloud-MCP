# Docker MCP Registry: Volume Mount Specification

## Registry server.yaml Schema

The Docker MCP Registry `server.yaml` supports a `run` section with volume declarations:

```yaml
run:
  command: []string          # Override container entrypoint/arguments
  volumes: []string          # Mount host paths into the container
  user: string               # Container user (UID or name)
  env: map[string]string     # Static runtime environment variables
  allowHosts: []string       # Egress firewall rules (domain:port)
  disableNetwork: bool       # Disable all network access
```

## Current server.yaml (No Mounts)

**File:** `registry/server.yaml`

```yaml
name: aspose-barcode-cloud
image: mcp/aspose-barcode-cloud
type: server
config:
  secrets:
    - name: aspose-barcode.client_id
      env: ASPOSE_CLOUD_CLIENT_ID
    - name: aspose-barcode.client_secret
      env: ASPOSE_CLOUD_CLIENT_SECRET
```

No `run` section exists - no volumes, no command overrides.

## Reference: Filesystem MCP Server (Docker Registry Pattern)

The official filesystem MCP server is the canonical example of mount usage:

```yaml
# servers/filesystem/server.yaml
run:
  command:
    - "{{filesystem.paths|volume-target|into}}"
  volumes:
    - "{{filesystem.paths|volume|into}}"
  disableNetwork: true
config:
  parameters:
    type: object
    properties:
      paths:
        type: array
        items:
          type: string
        default:
          - /Users/local-test
    required:
      - paths
```

## Template Syntax for Volumes

| Pattern | Meaning |
|---------|---------|
| `{{server.param}}` | Single string parameter substitution |
| `{{server.param}}:/container/path` | Map host path to fixed container path |
| `{{server.param\|volume\|into}}` | Expand array into multiple `-v host:host` bind mounts |
| `{{server.param\|volume-target\|into}}` | Expand array into container-side target paths (for command args) |

## Other Registry Examples with Mounts

**Git server:**
```yaml
run:
  volumes:
    - "{{git.paths|volume|into}}"
```

**AKS server (specific path mappings):**
```yaml
run:
  volumes:
    - "{{aks.azure_dir}}:/home/mcp/.azure"
    - "{{aks.kubeconfig}}:/home/mcp/.kube/config"
  user: "{{aks.container_user}}"
```

**Docker CLI:**
```yaml
volumes:
  - /var/run/docker.sock:/var/run/docker.sock
```

## How MCP Clients Handle Docker Mounts

### Claude Desktop (claude_desktop_config.json)

No native `volumes` field. Mounts must be in Docker CLI args:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "--mount", "type=bind,src=/Users/username/Desktop,dst=/projects/Desktop",
        "mcp/filesystem",
        "/projects/Desktop"
      ]
    }
  }
}
```

When using **Docker MCP Toolkit** (integrated into Docker Desktop), the Toolkit provides a management UI with default-deny filesystem model - users explicitly select which servers receive mounts.

### VS Code (.vscode/mcp.json)

Also no native volume field. Uses Docker args with `${workspaceFolder}` substitution:

```json
{
  "servers": {
    "myserver": {
      "type": "stdio",
      "command": "docker",
      "args": ["run", "-i", "--rm", "-v", "${workspaceFolder}:/workspace", "mcp/myserver"]
    }
  }
}
```

### Docker MCP Gateway (gordon-mcp.yml)

Uses Docker Compose syntax:
```yaml
services:
  fs:
    image: mcp/filesystem
    command:
      - /rootfs
    volumes:
      - .:/rootfs
```

## Documentation References

- Docker MCP Registry: https://github.com/docker/mcp-registry
- Docker MCP Catalog docs: https://docs.docker.com/ai/mcp-catalog-and-toolkit/catalog/
- Docker MCP Toolkit docs: https://docs.docker.com/ai/mcp-catalog-and-toolkit/toolkit/
- Docker MCP Gateway: https://github.com/docker/mcp-gateway
- MCP Filesystem Server: https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem
