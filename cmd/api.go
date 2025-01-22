package main

import (
	"embed"
	"fmt"
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
	router.GET("/hfstyle", hochfrequenzStylesheetHandler)
	router.GET("/yanone-kaffeesatz-bold", yanoneKaffeesatzBoldHandler)
	router.GET("/roboto-condensed-regular", robotoCondensedRegularHandler)
	router.GET("/logo", logoHandler)
	router.GET("/favicon", faviconHandler)

	return router
}

// getIdGenerator checks the environment variables and decides which IdGenerator to use.
// this is useful if you want to use the same code base and deploy it to different environments (with different env variables) for different ID types
func getIdGenerator() (IdGenerator, error) {
	// set this value in local.settings.json or in the azure portal function settings
	if idTypeToGenerate, ok := os.LookupEnv("ID_TYPE_TO_GENERATE"); ok {
		idTypeToGenerate = strings.ToUpper(idTypeToGenerate)
		if idTypeToGenerate == "MALO" {
			return MaLoIdGenerator{}, nil
		}
		if idTypeToGenerate == "NELO" {
			return NeLoIdGenerator{}, nil
		}
		if idTypeToGenerate == "MELO" {
			return MeLoIdGenerator{}, nil
		}
		if idTypeToGenerate == "TRID" {
			return TRIdGenerator{}, nil
		}
		if idTypeToGenerate == "SRID" {
			return SRIdGenerator{}, nil
		}
		return nil, fmt.Errorf("unsupported value of environment variable 'ID_TYPE_TO_GENERATE': '%s'. Supported values are 'MALO', 'NELO', 'MELO', .", idTypeToGenerate)
	}
	return nil, fmt.Errorf("no value set for environment variable 'ID_TYPE_TO_GENERATE'. Supported values are 'MALO', 'NELO', 'MELO', 'TRID' and 'SRID'.")
}

func generateRandomId(c *gin.Context) {
	generator, err := getIdGenerator()
	if err != nil {
		c.JSON(501, gin.H{"error": err.Error()})
		return
	}
	generator.GenerateId(c)
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

//go:embed static/companystylesheet/css/hochfrequenz.css
var hochfrequenzStylesheet embed.FS

//go:embed static/companystylesheet/fonts/Roboto_Condensed/RobotoCondensed-Regular.ttf
var robotoRegularFont embed.FS

//go:embed static/companystylesheet/fonts/YanoneKaffeesatzTTF/YanoneKaffeesatz-Bold.ttf
var yanoneBoldFont embed.FS

//go:embed static/companystylesheet/logo.svg
var hfLogo embed.FS

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

// returns the hochfrequenz stylesheet as text/css
func hochfrequenzStylesheetHandler(c *gin.Context) {
	stylesheetBody, err := hochfrequenzStylesheet.ReadFile("static/companystylesheet/css/hochfrequenz.css")
	if err != nil {
		response := map[string]string{}
		c.JSON(http.StatusNotFound, response)
	}
	c.Data(http.StatusOK, "text/css", stylesheetBody)
}

// returns the yanone kaffeesatz (bold) ttf
func yanoneKaffeesatzBoldHandler(c *gin.Context) {
	ttfBody, err := yanoneBoldFont.ReadFile("static/companystylesheet/fonts/YanoneKaffeesatzTTF/YanoneKaffeesatz-Bold.ttf")
	if err != nil {
		response := map[string]string{}
		c.JSON(http.StatusNotFound, response)
	}
	c.Data(http.StatusOK, "font/ttf", ttfBody)
}

func robotoCondensedRegularHandler(c *gin.Context) {
	ttfBody, err := robotoRegularFont.ReadFile("static/companystylesheet/fonts/Roboto_Condensed/RobotoCondensed-Regular.ttf")
	if err != nil {
		response := map[string]string{}
		c.JSON(http.StatusNotFound, response)
	}
	c.Data(http.StatusOK, "font/ttf", ttfBody)
}

// returns the hochfrequenz logo as image/svg+xml
func logoHandler(c *gin.Context) {
	ttfBody, err := hfLogo.ReadFile("static/companystylesheet/logo.svg")
	if err != nil {
		response := map[string]string{}
		c.JSON(http.StatusNotFound, response)
	}
	c.Data(http.StatusOK, "image/svg+xml", ttfBody)
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
