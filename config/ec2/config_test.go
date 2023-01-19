package ec2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	configPath := "./config_test.yaml"
	defaultUrl := "169.254.169.254"

	ec2Config, err := GetConfig(configPath)
	assert.Nil(t, err)

	assert.NotEqual(t, defaultUrl, ec2Config.V1Url)
	assert.NotEqual(t, defaultUrl, ec2Config.V2Url)
	assert.True(t, ec2Config.BasicAuth.Enabled)
	assert.NotEqual(t, "", ec2Config.BasicAuth.Username)
	assert.NotEqual(t, "", ec2Config.BasicAuth.Password)
	assert.Contains(t, ec2Config.AllowPathPrefixes, "/latest/api/token")

}
