package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPServerExposesAllIDGeneratorsOverHTTP(t *testing.T) {
	httpServer := httptest.NewServer(newMCPHandler())
	defer httpServer.Close()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "malo-id-generator-test-client",
		Version: "1.0.0",
	}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint: httpServer.URL,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	expectedTools := map[string]bool{
		"generate_malo_id":  false,
		"generate_nelo_id":  false,
		"generate_melo_id":  false,
		"generate_tr_id":    false,
		"generate_sr_id":    false,
		"generate_lobue_id": false,
	}
	for _, tool := range tools.Tools {
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		}
	}
	for name, found := range expectedTools {
		if !found {
			t.Errorf("MCP tool %q was not registered", name)
		}
	}

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "generate_malo_id",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatal("generate_malo_id returned an MCP tool error")
	}
}
