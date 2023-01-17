package ec2

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	// default
	_, err := GetConfig()
	assert.NotNil(t, err)

	// user of password not set
	os.Setenv("IMDS_BASIC_AUTH_USERNAME", "test-user")
	os.Setenv("IMDS_BASIC_AUTH_PASSWORD", "")
	_, err = GetConfig()
	assert.NotNil(t, err)

	// user of password not set
	os.Setenv("IMDS_BASIC_AUTH_USERNAME", "")
	os.Setenv("IMDS_BASIC_AUTH_PASSWORD", "test-pass")
	_, err = GetConfig()
	assert.NotNil(t, err)

	// success(auth disabled)
	os.Setenv("IMDS_BASIC_AUTH_USERNAME", "")
	os.Setenv("IMDS_BASIC_AUTH_PASSWORD", "")
	os.Setenv("IMDS_BASIC_AUTH_ENABLED", "false")
	_, err = GetConfig()
	assert.Nil(t, err)

	// success(auth enabled)
	os.Setenv("IMDS_BASIC_AUTH_USERNAME", "test-user")
	os.Setenv("IMDS_BASIC_AUTH_PASSWORD", "test-pass")
	os.Setenv("IMDS_BASIC_AUTH_ENABLED", "true")

	_, err = GetConfig()
	assert.Nil(t, err)

}
