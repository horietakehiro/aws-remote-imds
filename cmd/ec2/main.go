package main

import (
	"log"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	ec2Config "aws-remote-imds/config/ec2"
)

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
	}))
	gv2.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: middleware.NewRoundRobinBalancer(v2Targets),
		Rewrite: map[string]string{
			"^/imds/v2/*": "/$1",
		},
	}))

	listenAddress := config.String("listenAddress")
	log.Printf("use %s as listen address", listenAddress)
	e.Logger.Fatal(e.Start(listenAddress))

}
