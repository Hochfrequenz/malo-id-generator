package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
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
	// the following pathes have to match the name of the respective azure function or its route (if set, e.g. in case of function generate-malo-id whose route in function.json is "/")
	// see this SO answer: https://stackoverflow.com/a/76419027/10009545
	router.GET("/", generateRandomId)
	router.GET("/style", stylesheetHandler)
	router.GET("/favicon", faviconHandler)

	return router
}

func getIdGenerator() IdGenerator {
	// todo: check env variable on which IdGenerator to load
	return MaLoIdGenerator{}
}

func generateRandomId(c *gin.Context) {
	maloIdGenerator := getIdGenerator()
	maloIdGenerator.GenerateId(c)
}

func getPort() string {
	port := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		port = ":" + val
	}
	return port
}

// embedded files:
//
//go:embed static/style.css
var stylesheet embed.FS

// favicon is the favicon (the little icon in the browser tab)
//
//go:embed static/favicon.png
var favicon embed.FS

// templatesFS is the embedded file system where the template files for gin are located
//
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
