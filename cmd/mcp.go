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

	addMCPTool(server, "generate_malo_id",
		"Marktlokations-ID (MaLo)",
		"Generiert eine zufällige, 11-stellige Marktlokations-ID (MaLo) mit gültiger Prüfziffer gemäß BDEW-Konvention, bestehend aus einer 10-stelligen Zufallszahl (ohne führende Null) und einer einstelligen Prüfziffer. Das Ergebnis enthält zusätzlich die Prüfziffer separat, die ID ohne Prüfziffer und die Vergabestelle (DVGW oder BDEW, abhängig von der ersten Ziffer). Nützlich für Testdaten und Demo-Systeme im Kontext der Marktkommunikation in der Energiewirtschaft.",
		MaLoIdGenerator{})
	addMCPTool(server, "generate_nelo_id",
		"Netzlokations-ID (NeLo)",
		"Generiert eine zufällige, 11-stellige Netzlokations-ID (NeLo) mit gültiger Prüfziffer, bestehend aus einem führenden 'E', neun zufälligen alphanumerischen Zeichen und einer einstelligen Prüfziffer. Das Ergebnis enthält zusätzlich die Prüfziffer separat und die ID ohne Prüfziffer. Nützlich für Testdaten und Demo-Systeme im Kontext der Marktkommunikation in der Energiewirtschaft.",
		NeLoIdGenerator{})
	addMCPTool(server, "generate_melo_id",
		"Messlokations-ID (MeLo)",
		"Generiert eine zufällige, 33-stellige Messlokations-ID (MeLo) gemäß VDE-AR-N 4400, aufgebaut aus der Länderkennung 'DE', einer 6-stelligen (zufälligen, im Allgemeinen ungültigen) Netzbetreibernummer, einer 5-stelligen (zufälligen, im Allgemeinen ungültigen) Postleitzahl und 20 zufälligen alphanumerischen Zeichen als laufende Nummer. MeLo-IDs haben keine Prüfziffer. Das Ergebnis enthält die Bestandteile zusätzlich einzeln. Nützlich für Testdaten und Demo-Systeme im Kontext der Marktkommunikation in der Energiewirtschaft.",
		MeLoIdGenerator{})
	addMCPTool(server, "generate_tr_id",
		"Technische Ressourcen-ID (TR)",
		"Generiert eine zufällige, 11-stellige Technische Ressourcen-ID (TR-ID) mit gültiger Prüfziffer, bestehend aus einem führenden 'D', neun zufälligen alphanumerischen Zeichen und einer einstelligen Prüfziffer. Das Ergebnis enthält zusätzlich die Prüfziffer separat und die ID ohne Prüfziffer. Nützlich für Testdaten und Demo-Systeme im Kontext der Marktkommunikation in der Energiewirtschaft.",
		TRIdGenerator{})
	addMCPTool(server, "generate_sr_id",
		"Steuerbare Ressourcen-ID (SR)",
		"Generiert eine zufällige, 11-stellige Steuerbare Ressourcen-ID (SR-ID) mit gültiger Prüfziffer, bestehend aus einem führenden 'C', neun zufälligen alphanumerischen Zeichen und einer einstelligen Prüfziffer. Das Ergebnis enthält zusätzlich die Prüfziffer separat und die ID ohne Prüfziffer. Nützlich für Testdaten und Demo-Systeme im Kontext der Marktkommunikation in der Energiewirtschaft.",
		SRIdGenerator{})
	addMCPTool(server, "generate_lobue_id",
		"Lokationsbündel-ID (LoBü)",
		"Generiert eine zufällige, 11-stellige Lokationsbündel-ID (LoBü-ID) mit gültiger Prüfziffer, bestehend aus einem führenden 'G', neun zufälligen alphanumerischen Zeichen und einer einstelligen Prüfziffer. Das Ergebnis enthält zusätzlich die Prüfziffer separat und die ID ohne Prüfziffer. Nützlich für Testdaten und Demo-Systeme im Kontext der Marktkommunikation in der Energiewirtschaft.",
		LoBueIdGenerator{})

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

func addMCPTool(server *mcp.Server, name, title, description string, generator IdGenerator) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        name,
		Title:       title,
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
