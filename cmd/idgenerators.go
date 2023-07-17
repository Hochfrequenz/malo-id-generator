package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hochfrequenz/go-bo4e/bo"
	"github.com/hochfrequenz/go-bo4e/enum/rollencodetyp"
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
	})
	log.Printf("Successfully generated the MaLo '%s'", maloId)
}
