# High-Level Design: Mount-Based File Distribution

## Overview

This document describes the target architecture for replacing base64 file encoding with filesystem-based file exchange via Docker volume mounts. The mount is **required** -- `ASPOSE_CLOUD_MOUNT_PATH` must be set or the server exits with an error on startup. Base64 encoding is fully removed from the codebase.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│  HOST MACHINE                                                       │
│                                                                     │
│  ┌──────────────┐         ┌──────────────────────────────────────┐  │
│  │  MCP Client   │  stdio  │  Docker Container                    │  │
│  │  (Claude,     │◄───────►│                                      │  │
│  │   VS Code,    │ JSON-RPC│  ┌────────────────────────────────┐  │  │
│  │   etc.)       │         │  │  MCP Server (Go binary)        │  │  │
│  └──────┬───────┘         │  │                                  │  │  │
│         │                  │  │  ASPOSE_CLOUD_MOUNT_PATH=       │  │  │
│         │                  │  │    /mnt/data                    │  │  │
│         │                  │  │  ┌──────────┐ ┌──────────────┐  │  │  │
│         │                  │  │  │ Generate  │ │ Recognize/   │  │  │  │
│         │                  │  │  │ Handler   │ │ Scan Handler │  │  │  │
│         │                  │  │  └─────┬────┘ └──────┬───────┘  │  │  │
│         │                  │  │        │write      read│         │  │  │
│         │                  │  │        ▼              ▼          │  │  │
│         │                  │  │  ┌──────────────────────────┐   │  │  │
│         │                  │  │  │   Path Validator         │   │  │  │
│         │                  │  │  │   (security layer)       │   │  │  │
│         │                  │  │  └────────────┬─────────────┘   │  │  │
│         │                  │  └───────────────┼────────────────┘  │  │
│         │                  │                  │                    │  │
│         │                  │     ┌────────────▼─────────────┐     │  │
│         │                  │     │   /mnt/data (container)  │     │  │
│         │                  │     └────────────┬─────────────┘     │  │
│         │                  └──────────────────┼──────────────────┘  │
│         │                          bind mount │                     │
│         │                  ┌──────────────────▼──────────────────┐  │
│         └─────────────────►│   ~/barcode-data (host)             │  │
│           file access      │   qr-20260326-143022-a1b2c3d4.png  │  │
│                            │   code128-20260326-143023-e5f6.jpeg │  │
│                            └─────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

## Data Flow: Generate Barcode

```
MCP Client                    MCP Server                    Aspose Cloud API
    │                              │                              │
    │  CallTool(generate_barcode)  │                              │
    │  {type: "QR", data: "hi"}   │                              │
    │─────────────────────────────►│                              │
    │                              │  Generate(QR, "hi", opts)    │
    │                              │─────────────────────────────►│
    │                              │                              │
    │                              │◄─────────────────────────────│
    │                              │  []byte (image data)         │
    │                              │                              │
    │                              │  1. Generate filename:       │
    │                              │     qr-20260326-143022-      │
    │                              │     a1b2c3d4.png             │
    │                              │                              │
    │                              │  2. Validate path is within  │
    │                              │     mount directory          │
    │                              │                              │
    │                              │  3. os.WriteFile(            │
    │                              │     /mnt/data/qr-...-a1b2.   │
    │                              │     png, imageBytes, 0644)   │
    │                              │                              │
    │  CallToolResult:             │                              │
    │  TextContent{                │                              │
    │    "Generated barcode:       │                              │
    │     qr-...-a1b2c3d4.png     │                              │
    │     Format: image/png"       │                              │
    │  }                           │                              │
    │◄─────────────────────────────│                              │
    │                              │                              │
    │  Read file from host mount:  │                              │
    │  ~/barcode-data/qr-..-.png  │                              │
    │                              │                              │
```

**Note:** SVG format also writes to file and returns the path (same flow as raster formats per ADR-10).

## Data Flow: Recognize/Scan Barcode

```
MCP Client                    MCP Server                    Aspose Cloud API
    │                              │                              │
    │  1. Place image file in      │                              │
    │     ~/barcode-data/scan.png  │                              │
    │                              │                              │
    │  CallTool(scan_barcode)      │                              │
    │  {image_path: "scan.png"}    │                              │
    │─────────────────────────────►│                              │
    │                              │  1. Validate path            │
    │                              │  2. Resolve: /mnt/data/      │
    │                              │     scan.png                 │
    │                              │  3. os.Open(resolved)        │
    │                              │                              │
    │                              │  ScanMultipart(ctx, file)    │
    │                              │─────────────────────────────►│
    │                              │◄─────────────────────────────│
    │                              │  BarcodeResponseList         │
    │                              │                              │
    │  CallToolResult:             │                              │
    │  TextContent{                │                              │
    │    "Found 1 barcode(s):      │                              │
    │     1. Type: QR              │                              │
    │        Value: hello"         │                              │
    │  }                           │                              │
    │◄─────────────────────────────│                              │
```

## Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    main.go                               │
│  - Reads ASPOSE_CLOUD_MOUNT_PATH env var                │
│  - Exits with error if not set                          │
│  - Passes MountConfig to handler factories              │
└────────────────────────┬────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
┌────────────────┐ ┌──────────────┐ ┌──────────────┐
│ tool_generate  │ │tool_recognize│ │  tool_scan   │
│                │ │              │ │              │
│ MakeGenerate   │ │MakeRecognize │ │ MakeScan     │
│ Handler(client,│ │Handler(client│ │ Handler(     │
│   mount)       │ │  mount)      │ │client,mount) │
└───────┬────────┘ └──────┬───────┘ └──────┬───────┘
        │                 │                │
        ▼                 ▼                ▼
┌──────────────────────────────────────────────────────┐
│              mount.go (NEW FILE)                      │
│                                                       │
│  MountConfig struct {                                 │
│      Path string   // mount directory path            │
│  }                                                    │
│                                                       │
│  func NewMountConfig(envPath string) *MountConfig     │
│  func (m *MountConfig) GenerateFilename(              │
│      barcodeType, ext string) string                  │
│  func (m *MountConfig) WriteFile(                     │
│      filename string, data []byte) (string, error)    │
│  func (m *MountConfig) OpenFile(                      │
│      relativePath string) (*os.File, error)           │
│  func (m *MountConfig) ValidatePath(                  │
│      relativePath string) (string, error)             │
└──────────────────────────────────────────────────────┘
```

## Startup Logic

```
startup:
  mountPath = os.Getenv("ASPOSE_CLOUD_MOUNT_PATH")

  if mountPath == "":
    log.Fatalf("ASPOSE_CLOUD_MOUNT_PATH environment variable must be set")

  validate directory exists and is writable
  mountConfig = MountConfig{Path: mountPath}
  log.Printf("Mount mode enabled: %s", mountPath)

  pass mountConfig to all handler factories
```

## Handler Decision Tree

### generate_barcode handler:
```
receive imageBytes from Aspose API

ext = ExtensionForFormat(imageFormat)  // works for SVG too
filename = mount.GenerateFilename(barcodeType, ext)
mount.WriteFile(filename, imageBytes)
return TextContent with relative path + format metadata
```

### recognize_barcode handler:
```
validate input.ImagePath is not empty
validate image file extension
file = mount.OpenFile(input.ImagePath)
result = RecognizeMultipart(ctx, barcodeType, file, opts)
return TextContent with formatted results
```

### scan_barcode handler:
```
validate input.ImagePath is not empty
validate image file extension
file = mount.OpenFile(input.ImagePath)
result = ScanMultipart(ctx, file)
return TextContent with formatted results
```
