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
	// GenerateId generates and renders the result into the given gin context
	GenerateId(c *gin.Context)
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

// GenerateId of the MaLoIdGenerator returns a new random, 11 digit malo-id that has a valid check sum
func (m MaLoIdGenerator) GenerateId(c *gin.Context) {
	var maloIdWithoutChecksum string
	var maloCheckSum string
	for {
		maloIdWithoutChecksum = generateRandomString(allowedMaLoCharacters, 10)
		if maloIdWithoutChecksum[0] != '0' { // loop until he first character is not 0
			maloCheckSum = fmt.Sprintf("%d", bo.GetMaLoIdCheckSum(maloIdWithoutChecksum))
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
	c.HTML(http.StatusOK, "static/templates/malo.tmpl.html", gin.H{
		"maLoIdWithoutChecksum": maloIdWithoutChecksum,
		"checksum":              maloCheckSum,
		"issuer":                issuer.String(),
		"recruitingMessage":     template.HTML(recruitingMessage),
	})
	log.Printf("Successfully generated the MaLo '%s'", maloId)
}

// allowedNeLoCharacters contains those characters that are used to create new nelo ids
var allowedNeLoCharacters = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// NeLoIdGenerator is an IdGenerator that generates NeLo-IDs (Netzlokation-IDs)
type NeLoIdGenerator struct{}

// GenerateId of the NeLoIdGenerator returns a new random, 11 digit nelo-id that has a valid check sum
func (m NeLoIdGenerator) GenerateId(c *gin.Context) {
	var neloIdWithoutChecksum = "E" + generateRandomString(allowedNeLoCharacters, 9)
	var neloChecksum = fmt.Sprintf("%d", bo.GetNeLoIdCheckSum(neloIdWithoutChecksum))
	neloId := neloIdWithoutChecksum + neloChecksum
	c.HTML(http.StatusOK, "static/templates/nelo.tmpl.html", gin.H{
		"neLoIdWithoutChecksum": neloIdWithoutChecksum,
		"checksum":              neloChecksum,
		"recruitingMessage":     template.HTML(recruitingMessage),
	})
	log.Printf("Successfully generated the NeLo '%s'", neloId)
}
