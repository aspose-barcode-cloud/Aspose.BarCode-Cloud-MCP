# Aspose.BarCode Cloud MCP Server

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server that gives AI assistants the ability to generate, recognize, and scan barcodes using the [Aspose BarCode Cloud API](https://products.aspose.cloud/barcode/). Supports 60+ barcode symbologies including QR, Code128, DataMatrix, EAN, PDF417, Aztec, and more.

Barcode images are exchanged through a mounted data directory -- generated barcodes are saved as files, and images for recognition are read from the same directory.

## Prerequisites

- [Docker](https://www.docker.com/) (recommended) or [Go 1.25+](https://go.dev/)
- Aspose Cloud API credentials -- get them free at [dashboard.aspose.cloud](https://dashboard.aspose.cloud/applications)

## Quick Start

### 1. Get API Credentials

Sign up at [Aspose Cloud Dashboard](https://dashboard.aspose.cloud/applications) and create an application to get your **Client ID** and **Client Secret**.

### 2. Build the Docker Image

```bash
docker build -t aspose-barcode-cloud-mcp .
```

### 3. Configure Your MCP Client

Add the server to your MCP client configuration.

**Claude Desktop** (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "aspose-barcode-cloud": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "ASPOSE_CLOUD_CLIENT_ID",
        "-e", "ASPOSE_CLOUD_CLIENT_SECRET",
        "-v", "/path/to/your/barcode-data:/mnt/data",
        "aspose-barcode-cloud-mcp"
      ],
      "env": {
        "ASPOSE_CLOUD_CLIENT_ID": "your-client-id",
        "ASPOSE_CLOUD_CLIENT_SECRET": "your-client-secret"
      }
    }
  }
}
```

**VS Code** (`.vscode/mcp.json`):

```json
{
  "servers": {
    "aspose-barcode-cloud": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "ASPOSE_CLOUD_CLIENT_ID",
        "-e", "ASPOSE_CLOUD_CLIENT_SECRET",
        "-v", "${userHome}/barcode-data:/mnt/data",
        "aspose-barcode-cloud-mcp"
      ]
    }
  }
}
```

Replace `/path/to/your/barcode-data` with the directory where you want barcode files to be stored. Set the environment variables `ASPOSE_CLOUD_CLIENT_ID` and `ASPOSE_CLOUD_CLIENT_SECRET` in your shell or in the MCP client config.

> **Tip:** Create the mount directory yourself before starting the server (e.g. `mkdir -p ~/barcode-data`). If Docker creates it automatically, it will be owned by root and may cause permission issues.

## Tools

> **Note:** All `image_path` parameters are relative to the mounted data directory. For example, if your mount is `~/barcode-data` and the file is `~/barcode-data/photo.png`, pass `image_path` as `photo.png`.

The server exposes four MCP tools:

### `generate_barcode`

Generate a barcode image from text data.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `barcode_type` | Yes | Symbology to generate (e.g. `QR`, `Code128`, `DataMatrix`, `EAN13`) |
| `data` | Yes | Data to encode |
| `image_format` | No | Output format: `PNG` (default), `JPEG`, `SVG`, `GIF`, `TIFF` |
| `text_location` | No | Human-readable text position: `Below`, `Above`, `None` |
| `foreground_color` | No | Foreground color: name (e.g. `Black`) or `#AARRGGBB` hex |
| `background_color` | No | Background color: name (e.g. `White`) or `#AARRGGBB` hex |
| `resolution` | No | Image resolution in DPI |
| `rotation_angle` | No | Rotation: `0`, `90`, `180`, or `270` degrees |
| `image_width` | No | Image width in pixels |
| `image_height` | No | Image height in pixels |

### `recognize_barcode`

Recognize barcodes of a specific type from an image file.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `image_path` | Yes | Relative path to image in the mounted data directory (PNG, JPEG, GIF, TIFF, BMP) |
| `barcode_type` | No | Type to look for (default: most common types) |
| `recognition_mode` | No | Quality vs speed: `Fast`, `Normal`, `Excellent` |
| `recognition_image_kind` | No | Image source hint: `Photo`, `ScannedDocument`, `ClearImage` |

### `scan_barcode`

Auto-detect and read commonly used barcodes from an image -- no need to specify the barcode type.

| Parameter | Required | Description |
|-----------|----------|-------------|
| `image_path` | Yes | Relative path to image in the mounted data directory (PNG, JPEG, GIF, TIFF, BMP) |

### `list_barcode_types`

List all supported barcode symbologies for generation and recognition. No parameters required.

## Usage Examples

Once configured, ask your AI assistant to:

- *"Generate a QR code containing https://example.com"*
- *"Create a Code128 barcode with data ABC-123 in SVG format"*
- *"Scan the barcodes in image.png"*
- *"What barcode types are supported?"*

## Configuration

### CLI Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `--mount-path` | Yes | Absolute path to the data directory for file exchange |

> **Note:** When using Docker, the default mount path is `/mnt/data` (set via `CMD` in the Dockerfile). To use a custom path inside the container, pass it after the image name: `docker run ... aspose-barcode-cloud-mcp --mount-path=/custom/path`

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ASPOSE_CLOUD_CLIENT_ID` | Yes | Aspose Cloud application Client ID |
| `ASPOSE_CLOUD_CLIENT_SECRET` | Yes | Aspose Cloud application Client Secret |

## Building from Source

```bash
# Build
go build -o aspose-barcode-cloud-mcp .

# Run
export ASPOSE_CLOUD_CLIENT_ID="your-client-id"
export ASPOSE_CLOUD_CLIENT_SECRET="your-client-secret"
./aspose-barcode-cloud-mcp --mount-path="/path/to/barcode-data"
```

## Running Tests

```bash
# Unit tests (no credentials needed)
go test ./...

# Integration tests (requires credentials and network access)
export ASPOSE_CLOUD_CLIENT_ID="your-client-id"
export ASPOSE_CLOUD_CLIENT_SECRET="your-client-secret"
go test ./... -v
```

Integration tests are automatically skipped when credentials are not set or the Aspose Cloud API is unreachable.

## Supported Barcode Types

Over 60 symbologies including:

**1D:** Code128, Code39, Code93, EAN13, EAN8, UPCA, UPCE, Codabar, Interleaved2of5, MSI, Pharmacode, and more

**2D:** QR, DataMatrix, PDF417, Aztec, MaxiCode, DotCode, HanXin, and more

**Postal:** AustraliaPost, Postnet, Planet, RoyalMail, SingaporePost, and more

Use the `list_barcode_types` tool to see the full list.

## License

This project is licensed under the MIT License -- see the [LICENSE](LICENSE) file for details.
