package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
