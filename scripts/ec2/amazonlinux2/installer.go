package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

const (
	APP_NAME       = "aws-remote-imds"
	APP_DIR        = "/opt/aws-remote-imds/"
	SERVICE_DIR    = "/etc/systemd/system/"
	ARTIFACT_URL   = "https://public-artifact-bucket-382098889955-ap-northeast-1.s3.ap-northeast-1.amazonaws.com/aws-remote-imds/latest/amazonlinux2/amd64/artifacts.tar.gz"
	NGINX_CONF_DIR = "/ect/nginx/conf.d/"
	HTTPD_CONF_DIR = "/etc/httpd/conf.d/"

	// files in artifacts.tar.gz
	NGINX_CONF_FILE  = "nginx.conf"
	HTTPD_CONF_FILE  = "httpd.conf"
	SYSTEM_CONF_FILE = "system.service.template"
	APP_BIN_FILE     = "aws-remote-imds"
)

var (
	middleware string
	username   string
	password   string
)

func init() {
	flag.StringVar(&middleware, "m", "", "[required]middleware name that stands in front of this application. allowed values are 'nginx' or 'httpd'")
	flag.StringVar(&username, "u", "", "[required]username for basic auth")
	flag.StringVar(&password, "p", "", "[required]password for basic auth")
	flag.Parse()
	// custom validations
	isValid := validateArgs()
	if !isValid {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {

	execCommand([]string{"sudo", "mkdir", "-p", APP_DIR})
	execCommand([]string{"sudo", "chmod", "-R", "644", APP_DIR})
	execCommand([]string{"sudo", "chmod", "755", path.Join(APP_DIR, APP_BIN_FILE)})
	execCommand([]string{"sudo", "wget", "-P", APP_DIR, ARTIFACT_URL})
	artifactArchive := path.Join(APP_DIR, path.Base(ARTIFACT_URL))
	execCommand([]string{"sudo", "tar", "zxvf", artifactArchive})
	middleConfFile := path.Join(APP_DIR, fmt.Sprintf("%s.conf", middleware))
	execCommand([]string{
		"sudo", "cp", middleConfFile, fmt.Sprintf("/etc/%s/conf.d/%s.conf", middleware, APP_NAME),
	})
	serviceFile := path.Join(APP_DIR, fmt.Sprintf("%s.service", APP_NAME))
	execCommand([]string{
		"sudo", "cp", path.Join(APP_DIR, SYSTEM_CONF_FILE), serviceFile},
	)
	execCommand([]string{
		"sudo", "sed", "-i",
		"-e", fmt.Sprintf("\"s/_IMDS_BASIC_AUTH_USERNAME_/%s/g\"", username),
		"-e", fmt.Sprintf("\"s/IMDS_BASIC_AUTH_PASSWORD/%s/g\"", password),
		serviceFile,
	})
	execCommand([]string{"sudo", "mv", serviceFile, SERVICE_DIR})
	execCommand([]string{"sudo", "systemctl", "daemon-reload"})
	execCommand([]string{"sudo", "systemctl", "disable", APP_NAME})
	execCommand([]string{"sudo", "systemctl", "stop", APP_NAME})

	fmt.Println("==============================")
	fmt.Printf("You must manually start process: `systemctl start %s`\n", APP_NAME)
	fmt.Println("==============================")

}

func execCommand(command []string) {
	fmt.Printf("execute command$ %s\n", strings.Join(command, " "))
	out, err := exec.Command(command[0], command[1:]...).CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		panic(err)
	}
}

// func copyFile(src, dest string) error {
// 	srcFile, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer srcFile.Close()

// 	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	defer destFile.Close()

// 	_, err = io.Copy(destFile, srcFile)
// 	if err != nil {
// 		return err
// 	}

// 	return destFile.Close()

// }

func validateArgs() bool {
	fmt.Println(middleware)
	fmt.Println(username)
	fmt.Println(password)

	if middleware == "" {
		return false
	} else if middleware != "nginx" && middleware != "httpd" {
		return false
	}

	if username == "" || password == "" {
		return false
	}

	return true

}

// func downloadArtifact() (string, error) {
// 	res, err := http.Get(ARTIFACT_URL)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer res.Body.Close()

// 	localFile := path.Join(APP_DIR, "artifact.tar.gz")
// 	fileWriter, err := os.Create(localFile)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer fileWriter.Close()

// 	_, err = io.Copy(fileWriter, res.Body)
// 	if err != nil {
// 		return "", err
// 	}
// 	return localFile, nil

// }

// func decompress(arch string) (map[string]string, error) {
// 	var files = map[string]string{
// 		NGINX_CONF_FILE:  "",
// 		HTTPD_CONF_FILE:  "",
// 		SYSTEM_CONF_FILE: "",
// 		APP_BIN_FILE:     "",
// 	}
// 	file, err := os.Open(arch)
// 	if err != nil {
// 		return files, err
// 	}
// 	defer file.Close()

// 	gzipReader, err := gzip.NewReader(file)
// 	if err != nil {
// 		return files, err
// 	}
// 	tarReader := tar.NewReader(gzipReader)
// 	for {
// 		tarHeader, err := tarReader.Next()
// 		if err == io.EOF {
// 			break
// 		}
// 		if tarHeader.Name == NGINX_CONF_FILE {
// 			files[NGINX_CONF_FILE] = path.Join(APP_DIR, tarHeader.Name)
// 		}
// 		if tarHeader.Name == HTTPD_CONF_FILE {
// 			files[HTTPD_CONF_FILE] = path.Join(APP_DIR, tarHeader.Name)
// 		}
// 		if tarHeader.Name == SYSTEM_CONF_FILE {
// 			files[SYSTEM_CONF_FILE] = path.Join(APP_DIR, tarHeader.Name)
// 		}
// 		if tarHeader.Name == APP_BIN_FILE {
// 			files[APP_BIN_FILE] = path.Join(APP_DIR, tarHeader.Name)
// 		}
// 	}

// 	return files, nil

// }
