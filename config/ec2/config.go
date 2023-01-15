package ec2

import (
	"github.com/gookit/config/v2"
)

// GetConfig returns configs for ec2-imds reverse proxy server
func GetConfig() *config.Config {
	ec2Config := config.NewWithOptions("ec2Config", config.ParseEnv)

	// static values
	ec2Config.Set("listenAddress", ":9876")

	// environment variables with default values
	ec2Config.LoadOSEnvs(map[string]string{
		"IMDS_V1_URL": "v1Url",
		"IMDS_V2_URL": "v2Url",
	})
	if len(ec2Config.String("v1Url")) == 0 {
		ec2Config.Set("v1Url", "http://169.254.169.254")
	}
	if len(ec2Config.String("v2Url")) == 0 {
		ec2Config.Set("v2Url", "http://169.254.169.254")
	}

	return ec2Config
}
