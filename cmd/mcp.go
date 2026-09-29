package main

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpToolInput struct{}

func newMCPHandler() http.Handler {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "malo-id-generator",
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
	})
}

func addMCPTool(server *mcp.Server, name, description string, generator IdGenerator) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        name,
		Description: description,
	}, mcpToolHandler(generator))
}

func mcpToolHandler(generator IdGenerator) mcp.ToolHandlerFor[mcpToolInput, map[string]string] {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ mcpToolInput) (*mcp.CallToolResult, map[string]string, error) {
		return nil, generator.generateIdDictionary()
	}
}
