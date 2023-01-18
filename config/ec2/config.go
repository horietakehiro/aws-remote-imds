package ec2

import (
	"errors"
	"os"

	"github.com/gookit/config/v2"
)

func getEnvVal(key, defaultVal string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		return defaultVal
	}
	return val
}

// GetConfig returns configs for ec2-imds reverse proxy server
func GetConfig() (*config.Config, error) {
	ec2Config := config.NewWithOptions("ec2Config", config.ParseEnv)

	// environment variables with default values
	// ec2Config.Set("listenAddress", getEnvVal("IMDS_LISTEN_ADDRESS", ":9876"))
	ec2Config.Set("v1Url", getEnvVal("IMDS_V1_URL", "http://169.254.169.254"))
	ec2Config.Set("v2Url", getEnvVal("IMDS_V2_URL", "http://169.254.169.254"))
	ec2Config.Set("username", getEnvVal("IMDS_BASIC_AUTH_USERNAME", ""))
	ec2Config.Set("password", getEnvVal("IMDS_BASIC_AUTH_PASSWORD", ""))
	ec2Config.Set("basicAuthEnabled", getEnvVal("IMDS_BASIC_AUTH_ENABLED", "true"))

	// configuration assertions
	if ec2Config.Bool("basicAuthEnabled") && (ec2Config.String("username") == "" || ec2Config.String("password") == "") {
		return nil, errors.New("configuration error : if you enable basic auth(set IMDS_BASIC_AUTH_ENABLED 'true'), both username(IMDS_BASIC_AUTH_USERNAME) and password(IMDS_BASIC_AUTH_PASSWORD) must be specified as environment variable")
	}
	return ec2Config, nil
}
