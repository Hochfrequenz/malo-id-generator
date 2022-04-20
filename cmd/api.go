package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hochfrequenz/go-bo4e/bo"
	"html/template"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	router := NewRouter()
	err := router.Run(getPort())
	if err != nil {
		log.Panic(err)
	}
}

// NewRouter creates a gin engine and bind the handlers to the API paths
func NewRouter() *gin.Engine {
	router := gin.Default()
	pattern := "static/templates/*"
	loadHTMLFromEmbedFS(router, templatesFS, pattern)
	// router.LoadHTMLGlob("cmd/static/templates/*.html") // see https://gin-gonic.com/docs/examples/html-rendering/
	router.GET("/api/generate-malo-id", generateRandomMaLoId)
	router.GET("/api/style", stylesheetHandler)
	router.GET("/api/favicon", faviconHandler)
	router.GET("/api/logo", logoHandler)
	return router
}

func getPort() string {
	port := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		port = ":" + val
	}
	return port
}

// embedded files:
//go:embed static/style.css
var stylesheet embed.FS

// favicon is the favicon (the little icon in the browser tab)
//go:embed static/favicon.png
var favicon embed.FS

// logo is the hochfrequenz company logo
//go:embed static/logo.png
var logo embed.FS

// templatesFS is the embedded file system where the template files for gin are located
//go:embed static/templates
var templatesFS embed.FS

// returns the stylesheet as text/css
func stylesheetHandler(c *gin.Context) {
	stylesheetBody, err := stylesheet.ReadFile("static/style.css")
	if err != nil {
		response := map[string]string{}
		c.JSON(http.StatusNotFound, response)
	}
	c.Data(http.StatusOK, "text/css", stylesheetBody)
}

// returns the favicon as image/png
func faviconHandler(c *gin.Context) {
	stylesheetBody, err := favicon.ReadFile("static/favicon.png")
	if err != nil {
		response := map[string]string{}
		c.JSON(http.StatusNotFound, response)
	}
	c.Data(http.StatusOK, "image/png", stylesheetBody)
}

// returns the logo as image/png
func logoHandler(c *gin.Context) {
	stylesheetBody, err := logo.ReadFile("static/logo.png")
	if err != nil {
		response := map[string]string{}
		c.JSON(http.StatusNotFound, response)
	}
	c.Data(http.StatusOK, "image/png", stylesheetBody)
}

// allowedMaLoCharacters contains those characters that are used to create new malo ids
var allowedMaLoCharacters = []rune("123456789")

// generateRandomString returns a random combination of the allowed characters with given length
func generateRandomString(allowedCharacters []rune, length uint) string {
	// source: https://stackoverflow.com/a/22892986/10009545
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range b {
		b[i] = allowedCharacters[rand.Intn(len(allowedCharacters))]
	}
	time.Sleep(1 * time.Nanosecond) // gives the seed time to be refreshed
	return string(b)
}

// generateRandomMaLoId returns a new random, 11 digit malo-id that has a valid check sum and embeds
func generateRandomMaLoId(c *gin.Context) {
	var maloIdWithoutChecksum string
	var maloCheckSum string
	for {
		maloIdWithoutChecksum = generateRandomString(allowedMaLoCharacters, 10)
		if maloIdWithoutChecksum[0] != '0' { // loop until he first character is not 0
			maloCheckSum = fmt.Sprintf("%d", bo.GetMaLoIdCheckSum(maloIdWithoutChecksum))
			break
		}
	}
	maloId := maloIdWithoutChecksum + maloCheckSum
	c.HTML(http.StatusOK, "static/templates/index.tmpl.html", gin.H{
		"result": maloId,
	})
	log.Printf("Successfully generated the MaLo '%s'", maloId)
}

// boilerplate code to use embedded files as HTML templates:
// copied from here: https://github.com/gin-gonic/gin/issues/2795
// I don't care about linter warnings below this line

// nolint: goconst,gosimple
func loadHTMLFromEmbedFS(engine *gin.Engine, embedFS embed.FS, pattern string) {
	root := template.New("")
	tmpl := template.Must(root, loadAndAddToRoot(engine.FuncMap, root, embedFS, pattern))
	engine.SetHTMLTemplate(tmpl)
}

// nolint: goconst,gosimple
func loadAndAddToRoot(funcMap template.FuncMap, rootTemplate *template.Template, embedFS embed.FS, pattern string) error {
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")

	err := fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if matched, _ := regexp.MatchString(pattern, path); !d.IsDir() && matched {
			data, readErr := embedFS.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			t := rootTemplate.New(path).Funcs(funcMap)
			if _, parseErr := t.Parse(string(data)); parseErr != nil {
				return parseErr
			}
		}
		return nil
	})
	return err
}
