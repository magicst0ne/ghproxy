package main

import (
    kingpin "gopkg.in/alecthomas/kingpin.v2"
    "github.com/labstack/echo/v4"
    "net/http"
    "regexp"
    "strings"
    "fmt"
)

var (
    address = kingpin.Flag("address", "Listen Address").Default("0.0.0.0:5432").String()
    sslKey = kingpin.Flag("key", "Private Key").Default("server.key").String()
    sslCert = kingpin.Flag("cert", "Certificate").Default("server.pem").String()
)


func main() {

    kingpin.HelpFlag.Short('h')
    kingpin.Parse()

    expGhRelease := `^(?:https?://)?github\.com(?:/[^/\s]+)(?:/[^/\s]+)?/(?:releases|archive)(?:/[^/\s]+)*/?$`
    expGhBlob := `^(?:https?://)?github\.com(?:/[^/\s]+)(?:/[^/\s]+)?/(?:blob|raw)(?:/[^/\s]+)*/?$`
    expGhRaw := `^(?:https?://)?raw\.githubusercontent\.com(?:/[^/\s]+)(?:/[^/\s]+)(?:/[^/\s]+)*/?$`
    expGhGist := `^(?:https?://)?gist\.githubusercontent\.com(?:/[^/\s]+)(?:/[^/\s]+)(?:/[^/\s]+)*/?$`

    e := echo.New()
    e.GET("/gh/*", func(c echo.Context) error {

        realUrl := c.Request().URL.String()
        realUrl = realUrl[4:]

        if matchUrl(expGhRelease, realUrl) {
            return getRelease(c, realUrl)
        } else if matchUrl(expGhBlob, realUrl) {
            redirectUrl := strings.Replace(realUrl, "/blob/", "@", 1)
            redirectUrl = strings.Replace(redirectUrl, "github.com", "cdn.jsdelivr.net/gh", 1)
            return c.Redirect(302, redirectUrl)
            
        } else if matchUrl(expGhRaw, realUrl) {
            redirectUrl := strings.Replace(realUrl, "raw.githubusercontent.com", "cdn.jsdelivr.net/gh", 1)
            return c.Redirect(302, redirectUrl)
        } else if matchUrl(expGhGist, realUrl) {
            redirectUrl := strings.Replace(realUrl, "gist.githubusercontent.com", "cdn.jsdelivr.net/gh", 1)
            return c.Redirect(302, redirectUrl)
            
        } else {
            redirectUrl := strings.Replace(realUrl, "github.com", "cdn.jsdelivr.net/gh", 1)
            return c.Redirect(302, redirectUrl)
        }
    })
    e.GET("/", func(c echo.Context) error {
        return c.String(200, "It works")
    })
    e.Logger.Fatal(e.StartTLS(*address, *sslCert, *sslKey))
}

func getRelease(c echo.Context, requestURL string) error {

    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    req, err := http.NewRequest(http.MethodGet, requestURL, nil)
    if err != nil {
        return c.String(500, "Internal Server Error")
    }

    resp, err := client.Do(req)
    if err != nil {
        return c.String(500, "Internal Server Error")
    }

    if (resp.StatusCode==302) {
        redirectUrl := resp.Header.Get("Location")
        if (len(redirectUrl)>10) {
            return c.Redirect(302, redirectUrl)
        }
    }
    return c.String(404, "Not Found")
}

func matchUrl(exp string, url string) bool {
    matched, err := regexp.MatchString(exp, url) 
    if err != nil {
      fmt.Println(err.Error())
      return false
    }
    if (matched) {
        return true
    } else {
        return false
    }
}