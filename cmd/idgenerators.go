package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hochfrequenz/go-bo4e/bo"
	"github.com/hochfrequenz/go-bo4e/enum/rollencodetyp"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// An IdGenerator is something that can generate IDs. Typically those IDs are either Markt-, Mess- oder Netzlokation-IDs.
type IdGenerator interface {
	// GenerateId generates and renders the result into the given gin context as HTML
	GenerateId(c *gin.Context)
	// GenerateIdRaw renders a dictionary with the generated ID and the checksum but no surrounding HTML (a JSON response for technical rather than human users)
	GenerateIdRaw(c *gin.Context)
	// GenerateIdDictionary generates and returns a dictionary with the generated ID and some metadata
	generateIdDictionary() (map[string]string, error)
}

// recruitingMessage is a multi line HTML comment that is inserted into the rendered HTML page. It is defined here because for reasons unknown to me, it was always stripped from the parsed HTML template.
// See: https://stackoverflow.com/q/76707663/10009545
const recruitingMessage string = `<!--
  ________________________________________
< Hey, kennst du schon unsere Jobangebote? >
  ----------------------------------------
         \   ^__^
          \  (oo)\_______
             (__)\       )\/\
                 ||----w |
                 ||     ||
https://www.hochfrequenz.de/karriere/stellenangebote/full-stack-entwickler/
-->`

// allowedMaLoCharacters contains those characters that are used to create new malo ids
var allowedMaLoCharacters = []rune("0123456789")

// generateRandomString returns a random combination of the allowed characters with given length
func generateRandomString(allowedCharacters []rune, length uint) string {
	// source: https://stackoverflow.com/a/22892986/10009545
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, length)
	for i := range b {
		b[i] = allowedCharacters[r.Intn(len(allowedCharacters))]
	}
	time.Sleep(1 * time.Nanosecond) // gives the seed time to be refreshed
	return string(b)
}

// MaLoIdGenerator is an IdGenerator that generates MaLo-IDs (Marktlokations-IDs)
type MaLoIdGenerator struct{}

func (m MaLoIdGenerator) generateIdDictionary() (map[string]string, error) {
	var maloIdWithoutChecksum string
	var maloCheckSum string
	for {
		maloIdWithoutChecksum = generateRandomString(allowedMaLoCharacters, 10)
		if maloIdWithoutChecksum[0] != '0' { // loop until he first character is not 0
			maloCheckSumInt, err := bo.CalculateMaLoIdCheckSum(maloIdWithoutChecksum)
			if err != nil {
				return nil, err
			}
			maloCheckSum = fmt.Sprintf("%d", maloCheckSumInt)
			break
		}
	}
	var issuer rollencodetyp.Rollencodetyp
	// see https://bdew-codes.de/Content/Files/MaLo/2017-04-28-BDEW-Anwendungshilfe-MaLo-ID_Version1.0_FINAL.PDF
	if rune(maloIdWithoutChecksum[0]) < '4' {
		issuer = rollencodetyp.DVGW
	} else {
		issuer = rollencodetyp.BDEW
	}
	maloId := maloIdWithoutChecksum + maloCheckSum
	result := make(map[string]string)
	result["id"] = maloId
	result["maLoIdWithoutChecksum"] = maloIdWithoutChecksum
	result["checksum"] = maloCheckSum
	result["issuer"] = issuer.String()
	result["type"] = "MaLo"
	log.Printf("Successfully generated the MaLo '%s'", maloId)
	return result, nil
}
func (m MaLoIdGenerator) GenerateIdRaw(c *gin.Context) {
	rawId, err := m.generateIdDictionary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rawId)
}

// GenerateId of the MaLoIdGenerator returns a new random, 11 digit malo-id that has a valid check sum
func (m MaLoIdGenerator) GenerateId(c *gin.Context) {
	rawId, err := m.generateIdDictionary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "static/templates/malo.tmpl.html", gin.H{
		"maLoIdWithoutChecksum": rawId["maLoIdWithoutChecksum"],
		"checksum":              rawId["checksum"],
		"issuer":                rawId["issuer"],
		"recruitingMessage":     template.HTML(recruitingMessage),
	})
}

// allowedNeLoCharacters contains those characters that are used to create new nelo ids
var allowedNeLoCharacters = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// NeLoIdGenerator is an IdGenerator that generates NeLo-IDs (Netzlokation-IDs)
type NeLoIdGenerator struct{}

func (m NeLoIdGenerator) generateIdDictionary() (map[string]string, error) {
	var neloIdWithoutChecksum = "E" + generateRandomString(allowedNeLoCharacters, 9)
	_checksum, err := bo.GetNeLoIdCheckSum(neloIdWithoutChecksum)
	if err != nil {
		return nil, err
	}
	var neloChecksum = fmt.Sprintf("%d", _checksum)
	neloId := neloIdWithoutChecksum + neloChecksum
	result := make(map[string]string)
	result["id"] = neloId
	result["neLoIdWithoutChecksum"] = neloIdWithoutChecksum
	result["checksum"] = neloChecksum
	result["type"] = "NeLo"
	log.Printf("Successfully generated the NeLo '%s'", neloId)
	return result, nil
}

// GenerateId of the NeLoIdGenerator returns a new random, 11 digit nelo-id that has a valid check sum
func (m NeLoIdGenerator) GenerateId(c *gin.Context) {
	rawId, err := m.generateIdDictionary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "static/templates/nelo.tmpl.html", gin.H{
		"neLoIdWithoutChecksum": rawId["neLoIdWithoutChecksum"],
		"checksum":              rawId["checksum"],
		"recruitingMessage":     template.HTML(recruitingMessage),
	})
}
func (m NeLoIdGenerator) GenerateIdRaw(c *gin.Context) {
	rawId, err := m.generateIdDictionary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rawId)
}

// MeLoIdGenerator is an IdGenerator that generates MeLo-IDs (Messlokation-IDs)
type MeLoIdGenerator struct{}

var numbers = []rune("0123456789")
var allowedMeLoCharacters = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (m MeLoIdGenerator) generateIdDictionary() (map[string]string, error) {
	// See VDE-AR-N 4400 https://www.vde-verlag.de/normen/0400343/vde-ar-n-4400-anwendungsregel-2019-07.html
	/*              DE|001069|66646|10000000000000012345
	                 |     |      |        |
	  Landesziffern -|     |      |        |- Laufende Nummer
	                       |      |
	Netzbetreibernummer ---|      |-- PLZ
	*/
	const landesziffern = "DE"
	var netzbetreibernummer = generateRandomString(numbers, 6) // im Allgemeinen keine gültige ID
	var postleitzahl = generateRandomString(numbers, 5)        // im Allgemeinen nicht gültige PLZ
	var laufendeNummer = generateRandomString(allowedMeLoCharacters, 20)
	// 2+6+5+20 = 33
	var meloId = landesziffern + netzbetreibernummer + postleitzahl + laufendeNummer
	result := make(map[string]string)
	result["id"] = meloId
	result["landesziffern"] = landesziffern
	result["netzbetreibernummer"] = netzbetreibernummer
	result["postleitzahl"] = postleitzahl
	result["laufendeNummer"] = laufendeNummer
	result["type"] = "MeLo"
	log.Printf("Successfully generated the MeLo '%s'", meloId)
	return result, nil
}

// GenerateId of the MeLoIdGenerator returns a new random, 33 character melo-id; MeLo-IDs have no checksum
func (m MeLoIdGenerator) GenerateId(c *gin.Context) {
	rawId, err := m.generateIdDictionary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "static/templates/melo.tmpl.html", gin.H{
		"meloId":              rawId["id"],
		"landesziffern":       rawId["landesziffern"],
		"netzbetreibernummer": rawId["netzbetreibernummer"],
		"postleitzahl":        rawId["postleitzahl"],
		"laufendeNummer":      rawId["laufendeNummer"],
		"recruitingMessage":   template.HTML(recruitingMessage),
	})
}
func (m MeLoIdGenerator) GenerateIdRaw(c *gin.Context) {
	rawId, err := m.generateIdDictionary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rawId)
}

// Ressourcen-IDs

// allowedRessourcenIdCharacters contains those characters that are used to create new "Technische Ressourcen-IDs" and "Steuerbare Ressourcen-IDs"
var allowedRessourcenIdCharacters = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// TRIdGenerator is an IdGenerator that generates TR-IDs (Technische Ressourcen-IDs)
type TRIdGenerator struct{}

func (m TRIdGenerator) generateIdDictionary() (map[string]string, error) {
	var trIdWithoutChecksum = "D" + generateRandomString(allowedRessourcenIdCharacters, 9)
	_checksum, err := bo.GetTRIdCheckSum(trIdWithoutChecksum)
	if err != nil {
		return nil, err
	}
	var trIdChecksum = fmt.Sprintf("%d", _checksum)
	trId := trIdWithoutChecksum + trIdChecksum
	result := make(map[string]string)
	result["id"] = trId
	result["trIdWithoutChecksum"] = trIdWithoutChecksum
	result["checksum"] = trIdChecksum
	result["type"] = "TR"
	log.Printf("Successfully generated the TRID '%s'", trId)
	return result, nil
}

// GenerateId of the TRIdGenerator returns a new random, 11 digit tr-id that has a valid check sum
func (m TRIdGenerator) GenerateId(c *gin.Context) {
	rawId, err := m.generateIdDictionary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "static/templates/trid.tmpl.html", gin.H{
		"trIdWithoutChecksum": rawId["trIdWithoutChecksum"],
		"checksum":            rawId["checksum"],
		"recruitingMessage":   template.HTML(recruitingMessage),
	})
}
func (m TRIdGenerator) GenerateIdRaw(c *gin.Context) {
	rawId, err := m.generateIdDictionary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rawId)
}

// SRIdGenerator is an IdGenerator that generates SR-IDs (Steuerbare Ressourcen-IDs)
type SRIdGenerator struct{}

func (m SRIdGenerator) generateIdDictionary() (map[string]string, error) {
	var srIdWithoutChecksum = "C" + generateRandomString(allowedRessourcenIdCharacters, 9)
	_checksum, err := bo.GetSRIdCheckSum(srIdWithoutChecksum)
	if err != nil {
		return nil, err
	}
	var srIdChecksum = fmt.Sprintf("%d", _checksum)
	srId := srIdWithoutChecksum + srIdChecksum
	result := make(map[string]string)
	result["id"] = srId
	result["srIdWithoutChecksum"] = srIdWithoutChecksum
	result["checksum"] = srIdChecksum
	result["type"] = "SR"
	log.Printf("Successfully generated the SRID '%s'", srId)
	return result, nil
}

// GenerateId of the SRIdGenerator returns a new random, 11 digit sr-id that has a valid check sum
func (m SRIdGenerator) GenerateId(c *gin.Context) {
	rawId, err := m.generateIdDictionary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "static/templates/srid.tmpl.html", gin.H{
		"srIdWithoutChecksum": rawId["srIdWithoutChecksum"],
		"checksum":            rawId["checksum"],
		"recruitingMessage":   template.HTML(recruitingMessage),
	})
}
func (m SRIdGenerator) GenerateIdRaw(c *gin.Context) {
	rawId, err := m.generateIdDictionary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rawId)
}
