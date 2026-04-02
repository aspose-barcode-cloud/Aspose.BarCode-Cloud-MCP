# Task Summary

## Objective

Create an MCP (Model Context Protocol) server that wraps the **Aspose Barcode Cloud Go SDK** compatible for register it in the **Docker MCP Registry**.

## Input Artifacts

- Docker MCP Registry: https://github.com/docker/mcp-registry
- Contributing guide: https://github.com/docker/mcp-registry/blob/main/CONTRIBUTING.md
- Aspose Barcode Cloud Go SDK: https://github.com/aspose-barcode-cloud/aspose-barcode-cloud-go (v4, API v4.0)

## Deliverables

1. A Go MCP server exposing Aspose Barcode Cloud operations as MCP tools
2. A Dockerfile to containerize the server
3. Registry submission files (`server.yaml`, `tools.json`) conforming to Docker MCP Registry requirements
4. README/documentation for the server

## Key Constraints

- Must use Go language (the SDK is Go-native)
- Must run as a Docker container (local server type in the registry)
- License must be MIT (Aspose SDK is MIT)
- Must pass Docker MCP Registry CI validation
- Requires Aspose Cloud API credentials (Client ID + Client Secret) at runtime
