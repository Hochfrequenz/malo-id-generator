package main

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpToolInput struct{}

type mcpGeneratedID struct {
	ID                     string `json:"id"`
	Type                   string `json:"type"`
	Checksum               string `json:"checksum,omitempty"`
	Issuer                 string `json:"issuer,omitempty"`
	MaLoIDWithoutChecksum  string `json:"maLoIdWithoutChecksum,omitempty"`
	NeLoIDWithoutChecksum  string `json:"neLoIdWithoutChecksum,omitempty"`
	TRIDWithoutChecksum    string `json:"trIdWithoutChecksum,omitempty"`
	SRIDWithoutChecksum    string `json:"srIdWithoutChecksum,omitempty"`
	LoBueIDWithoutChecksum string `json:"loBueIdWithoutChecksum,omitempty"`
	Landesziffern          string `json:"landesziffern,omitempty"`
	Netzbetreibernummer    string `json:"netzbetreibernummer,omitempty"`
	Postleitzahl           string `json:"postleitzahl,omitempty"`
	LaufendeNummer         string `json:"laufendeNummer,omitempty"`
}

func newMCPHandler() http.Handler {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "id-generator",
		Version: "1.0.0",
	}, nil)

	addMCPTool(server, "generate_malo_id", "Generate a random valid MaLo-ID with checksum", MaLoIdGenerator{})
	addMCPTool(server, "generate_nelo_id", "Generate a random valid NeLo-ID with checksum", NeLoIdGenerator{})
	addMCPTool(server, "generate_melo_id", "Generate a random MeLo-ID", MeLoIdGenerator{})
	addMCPTool(server, "generate_tr_id", "Generate a random valid TR-ID with checksum", TRIdGenerator{})
	addMCPTool(server, "generate_sr_id", "Generate a random valid SR-ID with checksum", SRIdGenerator{})
	addMCPTool(server, "generate_lobue_id", "Generate a random valid LoBü-ID with checksum", LoBueIdGenerator{})

	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
		// behind the Azure Functions front end this custom handler process listens on a
		// loopback address while the request's Host header carries the public domain, which
		// the SDK's DNS rebinding protection (meant for localhost dev servers) would reject
		// with a 403 on every request; the public endpoint is served by Azure, not this
		// process, so the protection is meaningless here and must be disabled
		DisableLocalhostProtection: true,
	})
}

func addMCPTool(server *mcp.Server, name, description string, generator IdGenerator) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        name,
		Description: description,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, mcpToolHandler(generator))
}

func mcpToolHandler(generator IdGenerator) mcp.ToolHandlerFor[mcpToolInput, mcpGeneratedID] {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ mcpToolInput) (*mcp.CallToolResult, mcpGeneratedID, error) {
		generated, err := generator.generateIdDictionary()
		if err != nil {
			return nil, mcpGeneratedID{}, err
		}
		return nil, mcpGeneratedID{
			ID:                     generated["id"],
			Type:                   generated["type"],
			Checksum:               generated["checksum"],
			Issuer:                 generated["issuer"],
			MaLoIDWithoutChecksum:  generated["maLoIdWithoutChecksum"],
			NeLoIDWithoutChecksum:  generated["neLoIdWithoutChecksum"],
			TRIDWithoutChecksum:    generated["trIdWithoutChecksum"],
			SRIDWithoutChecksum:    generated["srIdWithoutChecksum"],
			LoBueIDWithoutChecksum: generated["loBueIdWithoutChecksum"],
			Landesziffern:          generated["landesziffern"],
			Netzbetreibernummer:    generated["netzbetreibernummer"],
			Postleitzahl:           generated["postleitzahl"],
			LaufendeNummer:         generated["laufendeNummer"],
		}, nil
	}
}
