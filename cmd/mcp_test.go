package main

import (
	"context"
	"net/http/httptest"
	"regexp"
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
	defer func() {
		if err := session.Close(); err != nil {
			t.Errorf("closing MCP session: %v", err)
		}
	}()

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
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Errorf("MCP tool %q is not marked read-only", tool.Name)
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
	generated, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("expected structured MCP result, got %T", result.StructuredContent)
	}
	id, ok := generated["id"].(string)
	if !ok || !regexp.MustCompile(`^[1-9][0-9]{10}$`).MatchString(id) {
		t.Errorf("expected an 11-digit MaLo-ID without a leading zero, got %v", generated["id"])
	}
	withoutChecksum, hasWithoutChecksum := generated["maLoIdWithoutChecksum"].(string)
	checksum, hasChecksum := generated["checksum"].(string)
	if !hasChecksum || !regexp.MustCompile(`^[0-9]$`).MatchString(checksum) {
		t.Errorf("expected a one-digit checksum, got %v", generated["checksum"])
	}
	if !hasWithoutChecksum || len(withoutChecksum) != 10 || id != withoutChecksum+checksum {
		t.Errorf("expected MaLo-ID and checksum fields to match, got %#v", generated)
	}
	if generated["type"] != "MaLo" {
		t.Errorf("expected type MaLo, got %v", generated["type"])
	}
	if generated["issuer"] != "DVGW" && generated["issuer"] != "BDEW" {
		t.Errorf("expected issuer DVGW or BDEW, got %v", generated["issuer"])
	}
	if len(generated) != 5 {
		t.Errorf("expected exactly the five MaLo result fields, got %#v", generated)
	}
}
