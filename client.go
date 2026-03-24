package main

import (
	"context"
	"fmt"

	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode/jwt"
)

// AsposeClient wraps the Aspose Barcode Cloud SDK client with authentication.
type AsposeClient struct {
	API     *barcode.APIClient
	AuthCtx context.Context
}

// NewAsposeClient creates a new authenticated Aspose Barcode Cloud client.
func NewAsposeClient(clientID, clientSecret string) (*AsposeClient, error) {
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("ASPOSE_CLIENT_ID and ASPOSE_CLIENT_SECRET must be set")
	}

	jwtConf := jwt.NewConfig(clientID, clientSecret)
	authCtx := context.WithValue(
		context.Background(),
		barcode.ContextJWT,
		jwtConf.TokenSource(context.Background()),
	)

	cfg := barcode.NewConfiguration()
	apiClient := barcode.NewAPIClient(cfg)

	return &AsposeClient{
		API:     apiClient,
		AuthCtx: authCtx,
	}, nil
}
