package main_test

import (
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/hochfrequenz/malo-id-generator/cmd"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
)

type Suite struct {
	suite.Suite
}

// SetupSuite sets up the tests
func (s *Suite) SetupSuite() {
}

func (s *Suite) AfterTest(_, _ string) {
}

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}

func performGetRequest(r http.Handler, path string) *httptest.ResponseRecorder {
	return performRequest(r, "GET", path, nil)
}

func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func (s *Suite) Test_Endpoint_Fails_Without_An_Environment_Variable() {
	err := os.Setenv("ID_TYPE_TO_GENERATE", "foobar") // set an unsupported value
	then.AssertThat(s.T(), err, is.Nil())
	router := main.NewRouter()
	response := performGetRequest(router, "/")
	then.AssertThat(s.T(), response.Code, is.EqualTo(http.StatusNotImplemented))
	responseBody := response.Body.String()
	then.AssertThat(s.T(), strings.Contains(responseBody, "ID_TYPE_TO_GENERATE"), is.True())
}

func (s *Suite) Test_MaLo_Endpoint_Returns_Something_Like_A_MaLo() {
	err := os.Setenv("ID_TYPE_TO_GENERATE", "malo")
	then.AssertThat(s.T(), err, is.Nil())
	maloPattern := regexp.MustCompile(`\d{10}<span [^>]+>\d</span>`)
	router := main.NewRouter()
	response := performGetRequest(router, "/")
	then.AssertThat(s.T(), response.Code, is.EqualTo(http.StatusOK))
	responseBody := response.Body.String()
	then.AssertThat(s.T(), maloPattern.MatchString(responseBody), is.True())
}

func (s *Suite) Test_NeLo_Endpoint_Returns_Something_Like_A_NeLo() {
	err := os.Setenv("ID_TYPE_TO_GENERATE", "nelo")
	then.AssertThat(s.T(), err, is.Nil())
	neloPattern := regexp.MustCompile(`E[A-Z\d]{9}<span [^>]+>\d</span>`)
	router := main.NewRouter()
	response := performGetRequest(router, "/")
	then.AssertThat(s.T(), response.Code, is.EqualTo(http.StatusOK))
	responseBody := response.Body.String()
	then.AssertThat(s.T(), neloPattern.MatchString(responseBody), is.True())
}

func (s *Suite) Test_MeLo_Endpoint_Returns_Something_Like_A_MeLo() {
	err := os.Setenv("ID_TYPE_TO_GENERATE", "melo")
	then.AssertThat(s.T(), err, is.Nil())
	router := main.NewRouter()
	response := performGetRequest(router, "/")
	then.AssertThat(s.T(), response.Code, is.EqualTo(http.StatusOK))
	responseBody := response.Body.String()
	containsDe := strings.Contains(responseBody, `<span class="landesziffern" title="Landescode (ISO 3166-1)">DE</span>`)
	// the test could be better, but it's just a quick check
	then.AssertThat(s.T(), containsDe, is.True())
}

func (s *Suite) Test_Stylesheet_Is_Returned() {
	router := main.NewRouter()
	response := performGetRequest(router, "/style")
	then.AssertThat(s.T(), response.Code, is.EqualTo(http.StatusOK))
}

func (s *Suite) Test_Favicon_Is_Returned() {
	router := main.NewRouter()
	response := performGetRequest(router, "/favicon")
	then.AssertThat(s.T(), response.Code, is.EqualTo(http.StatusOK))
}
