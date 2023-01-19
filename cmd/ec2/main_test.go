package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	ec2Config "aws-remote-imds/config/ec2"
)

func Test_main_v1_success(t *testing.T) {
	configYamlPath = "../../config/ec2/config_test.yaml"
	os.Setenv("IMDS_V1_URL", "http://localhost:1111")
	os.Setenv("IMDS_V2_URL", "http://localhost:2222")

	e := NewEchoServer(configYamlPath)
	s := httptest.NewServer(e.Server.Handler)
	defer e.Close()
	defer s.Close()

	t.Log(s.URL)

	config, _ := ec2Config.GetConfig(configYamlPath)
	for _, path := range config.AllowPathPrefixes {
		h := http.Client{}
		url, _ := url.JoinPath(s.URL, "imds", "v1", path)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.SetBasicAuth(config.BasicAuth.Username, config.BasicAuth.Password)
		res, _ := h.Do(req)
		assert.Equal(t, 200, res.StatusCode, "fail: %s", url)
		body := NewCustomBody()
		b, _ := io.ReadAll(res.Body)
		_ = json.Unmarshal(b, body)
		assert.Nil(t, body.InstanceMetadata.Error)
	}

}

func Test_main_v1_fail(t *testing.T) {
	configYamlPath = "../../config/ec2/config_test.yaml"
	os.Setenv("IMDS_V1_URL", "http://localhost:1111")
	os.Setenv("IMDS_V2_URL", "http://localhost:2222")

	e := NewEchoServer(configYamlPath)
	s := httptest.NewServer(e.Server.Handler)
	defer e.Close()
	defer s.Close()

	t.Log(s.URL)

	config, _ := ec2Config.GetConfig(configYamlPath)
	for _, path := range config.AllowPathPrefixes {
		h := http.Client{}
		url, _ := url.JoinPath(s.URL, "imds", "v1", path)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		// req.SetBasicAuth(config.BasicAuth.Username, config.BasicAuth.Password)
		res, _ := h.Do(req)
		assert.Equal(t, 401, res.StatusCode, "fail: %s", url)
	}

	h := http.Client{}
	url, _ := url.JoinPath(s.URL, "imds", "v1", "latest", "user-data")
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.SetBasicAuth(config.BasicAuth.Username, config.BasicAuth.Password)
	res, _ := h.Do(req)
	assert.NotEqual(t, 200, res.StatusCode, "fail: %s", url)

}

func Test_main_v2(t *testing.T) {
	configYamlPath = "../../config/ec2/config_test.yaml"

	os.Setenv("IMDS_V1_URL", "http://localhost:1111")
	os.Setenv("IMDS_V2_URL", "http://localhost:2222")

	e := NewEchoServer(configYamlPath)
	s := httptest.NewServer(e.Server.Handler)
	defer e.Close()
	defer s.Close()

	t.Log(s.URL)
	baseUrl, _ := url.JoinPath(s.URL, "imds", "v2")
	tokenUrl, _ := url.JoinPath(baseUrl, "latest/api/token")
	cookieJar, _ := cookiejar.New(nil)
	t.Logf("get token for imds v2 %s", tokenUrl)
	h := &http.Client{
		Jar: cookieJar,
	}
	req, _ := http.NewRequest(http.MethodPut, tokenUrl, nil)
	req.SetBasicAuth("test-user", "test-pass")
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "60")
	res, _ := h.Do(req)
	assert.Equal(t, 200, res.StatusCode, "fail: %s", tokenUrl)

	assert.NotEqual(t, "", res.Header.Get("X-aws-ec2-metadata-token"))
	assert.Equal(t, "60", res.Header.Get("X-aws-ec2-metadata-token-ttl-seconds"))

	newUrl, _ := url.JoinPath(baseUrl, "latest/meta-data/ami-id")
	req, _ = http.NewRequest(http.MethodGet, newUrl, nil)
	req.Header.Set("X-aws-ec2-metadata-token", res.Header.Get("X-aws-ec2-metadata-token"))
	req.SetBasicAuth("test-user", "test-pass")
	res, _ = h.Do(req)
	assert.Equal(t, 200, res.StatusCode, "fail: %s", newUrl)

}
