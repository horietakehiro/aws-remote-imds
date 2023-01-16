package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
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

func newCustomBody() CustomBody {
	return CustomBody{
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

func modifyResponse(r *http.Response) error {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	sBody := string(body)

	customBody := newCustomBody()

	// instance metadata
	customBody.InstanceMetadata.QueryPath = r.Request.URL.EscapedPath()
	if r.StatusCode == 200 {
		if strings.Contains(sBody, "\n") && !strings.Contains(r.Request.URL.EscapedPath(), "/user-data") {
			options := strings.Split(sBody, "\n")
			customBody.InstanceMetadata.Options = options[:len(options)-1]
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

	newBody, err := json.Marshal(customBody)
	if err != nil {
		return err
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(newBody))
	r.Header.Set("Content-Length", strconv.Itoa(len(newBody)))

	return nil
}

func main() {
	e := echo.New()
	config, err := ec2Config.GetConfig()
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	v1Url, err := url.Parse(config.String("v1Url"))
	if err != nil {
		e.Logger.Fatal(err)
	}
	log.Printf("use %s as imds v1 url", v1Url.String())

	v2Url, err := url.Parse(config.String("v2Url"))
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

	if config.Bool("basicAuthEnabled") {
		basicAuth := func(username, password string, ctx echo.Context) (bool, error) {
			if username == config.String("username") && password == config.String("password") {
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
	}))
	gv2.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: middleware.NewRoundRobinBalancer(v2Targets),
		Rewrite: map[string]string{
			"^/imds/v2/*": "/$1",
		},
		ModifyResponse: modifyResponse,
	}))

	listenAddress := config.String("listenAddress")
	log.Printf("use %s as listen address", listenAddress)
	e.Logger.Fatal(e.Start(listenAddress))

}
