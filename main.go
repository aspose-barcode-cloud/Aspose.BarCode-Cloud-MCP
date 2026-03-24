package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverVersion = "0.2604.0"

func main() {
	log.SetOutput(os.Stderr)

	// Read credentials from environment
	clientID := os.Getenv("ASPOSE_CLIENT_ID")
	clientSecret := os.Getenv("ASPOSE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatalf("ASPOSE_CLIENT_ID and ASPOSE_CLIENT_SECRET environment variables must be set")
	}

	// Create authenticated Aspose client
	client, err := NewAsposeClient(clientID, clientSecret)
	if err != nil {
		log.Fatalf("Failed to create Aspose client: %v", err)
	}

	log.Printf("Starting Aspose Barcode MCP Server v%s", serverVersion)

	// Create MCP server
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "aspose-barcode-cloud",
			Version: serverVersion,
		},
		nil,
	)

	// Register tools
	mcp.AddTool(s, &mcp.Tool{
		Name: "generate_barcode",
		Description: "Generate a barcode image of the specified type encoding the given data. " +
			"Returns the image as base64-encoded content. " +
			"Use list_barcode_types to see all supported barcode types.",
	}, makeGenerateHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name: "recognize_barcode",
		Description: "Recognize barcodes of a specific type from a base64-encoded image. " +
			"Allows specifying the barcode type and recognition quality. " +
			"For automatic detection of most commonly used barcode types, use scan_barcode instead or set MostCommonlyUsed barcode type.",
	}, makeRecognizeHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name: "scan_barcode",
		Description: "Automatically detect and read commonly used barcodes in a base64-encoded image. " +
			"Scans for most commonly used supported barcode types without requiring you to specify which type. " +
			"For targeted recognition of a specific barcode type, use recognize_barcode instead.",
	}, makeScanHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name: "list_barcode_types",
		Description: "List all supported barcode types for generation and recognition. " +
			"Use this to discover valid barcode_type values for generate_barcode and recognize_barcode.",
	}, makeListHandler())

	// Start stdio transport
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
