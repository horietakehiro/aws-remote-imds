package main

import (
	"encoding/json"
	"io"
	"net/http"
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
func Test_main(t *testing.T) {
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
