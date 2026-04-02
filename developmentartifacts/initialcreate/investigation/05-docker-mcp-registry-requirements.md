# Docker MCP Registry - Submission Requirements

## Overview

The Docker MCP Registry is a curated catalog of MCP servers discoverable via:
- Docker Hub MCP catalog (hub.docker.com/mcp)
- Docker Desktop MCP Toolkit
- Docker Hub's `mcp` namespace

## Server Types

**Local (our case)**: Containerized, Docker-built, runs in Docker containers.
**Remote**: Externally hosted, HTTP-based.

## Prerequisites

- Go v1.24+
- Docker Desktop
- Task (taskfile.dev)

## Submission File Structure

```
servers/aspose-barcode/
├── server.yaml       # Server configuration (required)
├── tools.json        # Tool definitions (required)
└── Dockerfile        # In the source repo, not in registry
```

## server.yaml Format (Local Server)

```yaml
name: aspose-barcode                     # unique server name
image: mcp/aspose-barcode                # Docker image (mcp/ prefix for Docker-built)
type: server                             # "server" for local
meta:
  category: productivity                 # category (see below)
  tags:
    - barcode
    - qr-code
    - image-generation
about:
  title: Aspose Barcode                  # human-readable title
  description: >-                        # clear description
    Generate, recognize, and scan barcodes using Aspose Barcode Cloud API.
  icon: https://...                      # icon URL
source:
  project: https://github.com/...        # source repo URL
  commit: <40-char-hash>                 # specific commit
config:
  description: Configure Aspose Cloud API credentials
  secrets:                               # sensitive values (user must provide)
    - name: aspose-barcode.client_id
      env: ASPOSE_CLOUD_CLIENT_ID
      example: <YOUR_CLIENT_ID>
    - name: aspose-barcode.client_secret
      env: ASPOSE_CLOUD_CLIENT_SECRET
      example: <YOUR_CLIENT_SECRET>
```

## tools.json Format

Array of tool definitions. Used when server can't list tools without configuration.

```json
[
  {
    "name": "generate_barcode",
    "description": "Generate a barcode image of specified type with given data",
    "arguments": [
      {
        "name": "barcode_type",
        "type": "string",
        "desc": "Barcode symbology (e.g., QR, Code128, DataMatrix)"
      },
      {
        "name": "data",
        "type": "string",
        "desc": "Data to encode in the barcode"
      }
    ]
  }
]
```

## Available Categories

From observed servers: database, devops, documentation, finance, productivity, etc.
Best fit for barcode server: **productivity** or **devops**

## Example server.yaml Files

### Stripe (local, with secrets)
```yaml
name: stripe
image: mcp/stripe
type: server
meta:
  category: finance
  tags:
    - stripe
    - finance
about:
  title: Stripe
  description: Interact with Stripe services over the Stripe API.
  icon: https://avatars.githubusercontent.com/u/856813?s=200&v=4
source:
  project: https://github.com/stripe/agent-toolkit
  commit: 5af4bcd15813cbcbd91baceeb5ec79cf975035f1
  directory: tools/modelcontextprotocol
run:
  command:
    - --tools=all
config:
  secrets:
    - name: stripe.secret_key
      env: STRIPE_SECRET_KEY
      example: sk_STRIPE_SECRET_KEY
```

### Puppeteer (local, minimal)
```yaml
name: puppeteer
image: mcp/puppeteer
type: server
meta:
  category: devops
  tags:
    - puppeteer
    - devops
about:
  title: Puppeteer (Archived)
  description: Browser automation and web scraping using Puppeteer.
  icon: https://avatars.githubusercontent.com/u/6906516?s=200&v=4
source:
  project: https://github.com/modelcontextprotocol/servers
  branch: 2025.4.24
  commit: 0123456789abcdef0123456789abcdef01234567
  dockerfile: src/puppeteer/Dockerfile
config:
  description: The MCP server is allowed to access these paths
  env:
    - name: DOCKER_CONTAINER
      example: "true"
      value: "true"
```

## Submission Process

1. Fork `docker/mcp-registry`
2. Generate server config via `task wizard` or `task create`
3. Build and test locally
4. Submit PR for Docker team review
5. Docker builds, signs, publishes the image with security features

## Licensing

- Contributions licensed under MIT
- Server's own license must allow public consumption (MIT/Apache 2 preferred)
