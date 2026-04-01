package main

import (
	"context"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/aspose-barcode-cloud/Aspose.BarCode-Cloud-MCP/mcpbarcode"
)

const serverVersion = "0.2604.0"

func main() {
	log.SetOutput(os.Stderr)

	// Read credentials from environment
	clientID := os.Getenv("ASPOSE_CLOUD_CLIENT_ID")
	clientSecret := os.Getenv("ASPOSE_CLOUD_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatalf("ASPOSE_CLOUD_CLIENT_ID and ASPOSE_CLOUD_CLIENT_SECRET environment variables must be set")
	}

	// Create authenticated Aspose client
	client, err := mcpbarcode.NewAsposeClient(clientID, clientSecret)
	if err != nil {
		log.Fatalf("Failed to create Aspose client: %v", err)
	}

	// Read mount path configuration (required)
	mountPath := os.Getenv("ASPOSE_CLOUD_MOUNT_PATH")
	mount, err := mcpbarcode.NewMountConfig(mountPath)
	if err != nil {
		log.Fatalf("Mount configuration error: %v", err)
	}
	log.Printf("Mount mode enabled: %s", mount.Path)

	// Create MCP server
	s := server.NewMCPServer("aspose-barcode-cloud", serverVersion)

	// Register tools
	s.AddTool(mcp.NewTool("generate_barcode",
		mcp.WithDescription("Generate a barcode image of the specified type encoding the given data. "+
			"Saves the image file to the mounted data directory and returns the file path. "+
			"Use list_barcode_types to see all supported barcode types."),
		mcp.WithInputSchema[mcpbarcode.GenerateBarcodeInput](),
	), mcpbarcode.MakeGenerateHandler(client, mount))

	s.AddTool(mcp.NewTool("recognize_barcode",
		mcp.WithDescription("Recognize barcodes of a specific type from an image file in the mounted data directory. "+
			"The image_path must be relative to the mounted directory. "+
			"Allows specifying the barcode type and recognition quality. "+
			"For automatic detection of most commonly used barcode types, use scan_barcode instead."),
		mcp.WithInputSchema[mcpbarcode.RecognizeBarcodeInput](),
	), mcpbarcode.MakeRecognizeHandler(client, mount))

	s.AddTool(mcp.NewTool("scan_barcode",
		mcp.WithDescription("Automatically detect and read commonly used barcodes from an image file "+
			"in the mounted data directory. The image_path must be relative to the mounted directory. "+
			"For targeted recognition of a specific barcode type, use recognize_barcode instead."),
		mcp.WithInputSchema[mcpbarcode.ScanBarcodeInput](),
	), mcpbarcode.MakeScanHandler(client, mount))

	s.AddTool(mcp.NewTool("list_barcode_types",
		mcp.WithDescription("List all supported barcode types for generation and recognition. "+
			"Use this to discover valid barcode_type values for generate_barcode and recognize_barcode."),
		mcp.WithInputSchema[mcpbarcode.ListBarcodeTypesInput](),
	), mcpbarcode.MakeListHandler())

	// Start stdio transport
	stdioServer := server.NewStdioServer(s)
	if err := stdioServer.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
