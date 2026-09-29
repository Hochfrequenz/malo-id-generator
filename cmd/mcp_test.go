package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
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

// Reproduces the production 403 (fixed via DisableLocalhostProtection): behind the Azure
// Functions front end the custom handler process listens on a loopback address while the
// request's Host header carries the public domain. The SDK's DNS rebinding protection
// (meant for localhost dev servers) rejects exactly this combination with a 403.
func TestMCPEndpointAcceptsPublicHostBehindLoopbackFrontEnd(t *testing.T) {
	handler := newMCPHandler()

	// httptest.NewServer listens on 127.0.0.1 (loopback) just like the handler process behind Azure
	ts := httptest.NewServer(handler)
	defer ts.Close()

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`
	req, _ := http.NewRequest("POST", ts.URL+"/mcp", strings.NewReader(body))
	req.Host = "technische.ressource.id" // the public domain Azure forwards the request for
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("closing response body: %v", err)
		}
	}()
	if resp.StatusCode == http.StatusForbidden {
		t.Fatalf("MCP endpoint returned 403 for public host behind loopback front end; " +
			"DisableLocalhostProtection is not effective")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from MCP endpoint, got %s", resp.Status)
	}
}
