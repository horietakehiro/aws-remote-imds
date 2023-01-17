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
)

func recursiveRequest(t *testing.T, baseUrl string, childPath []string) {
	for _, p := range childPath {
		newUrl, _ := url.JoinPath(baseUrl, p)
		t.Logf("test for %s", newUrl)

		h := http.Client{}
		req, _ := http.NewRequest(http.MethodGet, newUrl, nil)
		req.SetBasicAuth("test-user", "test-pass")
		res, _ := h.Do(req)

		assert.Equal(t, 200, res.StatusCode, "fail: %s", newUrl)

		body := NewCustomBody()
		b, _ := io.ReadAll(res.Body)
		_ = json.Unmarshal(b, body)

		assert.Nil(t, body.InstanceMetadata.Error)
		if len(body.InstanceMetadata.Options) > 0 {
			recursiveRequest(t, newUrl, body.InstanceMetadata.Options)
		}
	}
}

func Test_main_v1(t *testing.T) {
	os.Setenv("IMDS_V1_URL", "http://localhost:1111")
	os.Setenv("IMDS_V2_URL", "http://localhost:2222")
	os.Setenv("IMDS_BASIC_AUTH_USERNAME", "test-user")
	os.Setenv("IMDS_BASIC_AUTH_PASSWORD", "test-pass")

	e := NewEchoServer()
	s := httptest.NewServer(e.Server.Handler)
	defer e.Close()
	defer s.Close()

	t.Log(s.URL)
	initUrl, _ := url.JoinPath(s.URL, "imds")
	recursiveRequest(t, initUrl, []string{"v1"})

}

func Test_main_v2(t *testing.T) {
	os.Setenv("IMDS_V1_URL", "http://localhost:1111")
	os.Setenv("IMDS_V2_URL", "http://localhost:2222")
	os.Setenv("IMDS_BASIC_AUTH_USERNAME", "test-user")
	os.Setenv("IMDS_BASIC_AUTH_PASSWORD", "test-pass")

	e := NewEchoServer()
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

func Test_main_no_auth(t *testing.T) {
	os.Setenv("IMDS_V1_URL", "http://localhost:1111")
	os.Setenv("IMDS_V2_URL", "http://localhost:2222")
	os.Setenv("IMDS_BASIC_AUTH_ENABLED", "false")

	e := NewEchoServer()
	s := httptest.NewServer(e.Server.Handler)
	defer e.Close()
	defer s.Close()

	h := http.Client{}
	url, _ := url.JoinPath(s.URL, "imds/v1")
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	res, _ := h.Do(req)
	assert.Equal(t, 200, res.StatusCode, "fail: %s", url)

}
