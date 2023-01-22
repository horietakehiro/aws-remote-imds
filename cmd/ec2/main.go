package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"path"
	"time"

	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	ec2Config "aws-remote-imds/config/ec2"
)

type RequestMetadata struct {
	Proto           string   `json:"Proto"`
	XForwardedFor   []string `json:"X-Forwarded-For"`
	XRealIp         []string `json:"X-Real-Ip"`
	XForwardedProto []string `json:"X-Forwarded-Proto"`
}

type ResponseMetadata struct{}

type InstanceMetadata struct {
	QueryPath string   `json:"QueryPath"`
	Value     *string  `json:"Value"`
	Options   []string `json:"Options"`
	Error     *string  `json:"Error"`
}

type CustomBody struct {
	InstanceMetadata InstanceMetadata `json:"InstanceMetadata"`
	RequestMetadata  RequestMetadata  `json:"RequestMetadata"`
	ResponseMetadata ResponseMetadata `json:"ResponseMetadata"`
}

var (
	configYamlPath string
)

func NewCustomBody() *CustomBody {
	return &CustomBody{
		InstanceMetadata: InstanceMetadata{
			Value:   nil,
			Options: []string{},
			Error:   nil,
		},
		RequestMetadata: RequestMetadata{
			XForwardedFor:   []string{},
			XRealIp:         []string{},
			XForwardedProto: []string{},
		},
		ResponseMetadata: ResponseMetadata{},
	}
}

func requestSkipper(pathPrefix string, config ec2Config.Ec2Config) func(echo.Context) bool {
	return func(c echo.Context) bool {
		reqPath := c.Request().URL.EscapedPath()
		for _, p := range config.AllowPathPrefixes {
			// fullPath, _ := url.JoinPath(pathPrefix, path)
			fullPath := UrlJoinPath(pathPrefix, p)
			if strings.HasPrefix(reqPath, fullPath) {
				return false
			}
		}
		return true
	}
}

func modifyResponse(r *http.Response) error {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	sBody := string(body)

	customBody := NewCustomBody()

	// instance metadata
	customBody.InstanceMetadata.QueryPath = r.Request.URL.EscapedPath()
	if r.StatusCode == 200 {
		// in cases body contains multiline
		if strings.Contains(sBody, "\n") &&
			!strings.Contains(r.Request.URL.EscapedPath(), "/user-data") &&
			!strings.HasPrefix(sBody, "{") &&
			!strings.Contains(r.Request.URL.EscapedPath(), "/pkcs7") {

			options := strings.Split(sBody, "\n")
			customBody.InstanceMetadata.Options = options[:len(options)-1]

			// in cases body contains only 1 child path
		} else if strings.HasSuffix(sBody, "/") {
			customBody.InstanceMetadata.Options = []string{sBody}
		} else {
			customBody.InstanceMetadata.Value = &sBody
		}
	} else {
		customBody.InstanceMetadata.Error = &sBody
	}

	// request metadata
	if val := r.Request.Header.Get("X-Forwarded-For"); val != "" {
		customBody.RequestMetadata.XForwardedFor = strings.Split(val, ", ")
	}
	if val := r.Request.Header.Get("X-Forwarded-Proto"); val != "" {
		customBody.RequestMetadata.XForwardedProto = strings.Split(val, ", ")
	}
	if val := r.Request.Header.Get("X-Real-Ip"); val != "" {
		customBody.RequestMetadata.XRealIp = strings.Split(val, ", ")
	}
	customBody.RequestMetadata.Proto = r.Request.Proto

	if strings.Contains(r.Request.URL.EscapedPath(), "/api/token") && r.StatusCode == 200 {
		// set token as header
		r.Header.Set("X-aws-ec2-metadata-token", sBody)
		r.Header.Set(
			"X-aws-ec2-metadata-token-ttl-seconds",
			r.Request.Header.Get("X-aws-ec2-metadata-token-ttl-seconds"),
		)

		// set token as cookie
		expires, _ := strconv.Atoi(r.Request.Header.Get("X-aws-ec2-metadata-token-ttl-seconds"))
		cookies := http.Cookie{
			Name:    "X-aws-ec2-metadata-token",
			Value:   sBody,
			Domain:  r.Request.Host,
			Expires: time.Now().Add(time.Duration(expires) * time.Second),
		}
		r.Header.Set("Set-Cookie", cookies.String())
	}

	newBody, err := json.Marshal(customBody)
	if err != nil {
		return err
	}

	// replace response body
	r.Body = io.NopCloser(bytes.NewBuffer(newBody))
	r.Header.Set("Content-Length", strconv.Itoa(len(newBody)))
	r.Header.Set("Content-Type", "application/json")

	return nil
}

func UrlJoinPath(elems ...string) string {
	tmpUrl := path.Join(elems...)
	// convert http:/hogefuga... to http://hogefuga
	url := strings.Replace(tmpUrl, ":/", "://", 1)
	return url
}

func NewEchoServer(configPath string) *echo.Echo {
	e := echo.New()
	e.Pre(middleware.AddTrailingSlash())
	config, err := ec2Config.GetConfig(configYamlPath)
	if err != nil {
		e.Logger.Fatal(err)
	}
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	v1Url, err := url.Parse(config.V1Url)
	if err != nil {
		e.Logger.Fatal(err)
	}
	log.Printf("use %s as imds v1 url", v1Url.String())

	v2Url, err := url.Parse(config.V2Url)
	if err != nil {
		e.Logger.Fatal(err)
	}
	log.Printf("use %s as imds v2 url", v2Url.String())

	v1Targets := []*middleware.ProxyTarget{
		{
			URL: v1Url,
		},
	}
	v2Targets := []*middleware.ProxyTarget{
		{
			URL: v2Url,
		},
	}

	gv1 := e.Group("/imds/v1")
	gv2 := e.Group("/imds/v2")

	if config.BasicAuth.Enabled {
		basicAuth := func(username, password string, ctx echo.Context) (bool, error) {
			if username == config.BasicAuth.Username && password == config.BasicAuth.Password {
				return true, nil
			}
			return false, nil
		}
		gv1.Use(middleware.BasicAuth(basicAuth))
		gv2.Use(middleware.BasicAuth(basicAuth))
	}

	gv1.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: middleware.NewRandomBalancer(v1Targets),
		Rewrite: map[string]string{
			"^/imds/v1/*": "/$1",
		},
		ModifyResponse: modifyResponse,
		Skipper:        requestSkipper("/imds/v1/", config),
	}))
	gv2.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: middleware.NewRoundRobinBalancer(v2Targets),
		Rewrite: map[string]string{
			"^/imds/v2/*": "/$1",
		},
		ModifyResponse: modifyResponse,
		Skipper:        requestSkipper("/imds/v2/", config),
	}))

	return e

}

func init() {
	flag.StringVar(&configYamlPath, "f", "", "path to the config yaml file")
}

func main() {
	flag.Parse()

	e := NewEchoServer(configYamlPath)
	e.Logger.Fatal(e.Start(":9876"))
}
