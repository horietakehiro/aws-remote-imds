package ec2

import (
	"errors"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yamlv3"
)

// func getEnvVal(key, defaultVal string) string {
// 	val := os.Getenv(key)
// 	if len(val) == 0 {
// 		return defaultVal
// 	}
// 	return val
// }

type BasicAuthConfig struct {
	Enabled  bool   `mapstructure:"Enabled"`
	Username string `mapstructure:"Username"`
	Password string `mapstructure:"Password"`
}

type Ec2Config struct {
	V1Url      string          `mapstructure:"V1Url"`
	V2Url      string          `mapstructure:"V2Url"`
	BasicAuth  BasicAuthConfig `mapstructure:"BasicAuth"`
	AllowPaths []string        `mapstructure:"AllowPaths"`
}

// GetConfig returns configs for ec2-imds reverse proxy server
func GetConfig(filePath string) (Ec2Config, error) {

	ec2Config := Ec2Config{}

	config.WithOptions(config.ParseEnv)
	config.AddDriver(yamlv3.Driver)
	err := config.LoadFiles(filePath)
	if err != nil {
		return ec2Config, err
	}

	err = config.BindStruct("", &ec2Config)
	if err != nil {
		return ec2Config, err
	}

	// configuration assertions
	if ec2Config.BasicAuth.Enabled && (ec2Config.BasicAuth.Username == "" || ec2Config.BasicAuth.Password == "") {
		return ec2Config, errors.New("configuration error : if you enable basic auth, both username and password must be set")
	}

	return ec2Config, nil
}
