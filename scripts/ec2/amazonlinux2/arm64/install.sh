#!/bin/bash

APP_DIR="/opt/aws-remote-imds"
APP_NAME="ec2-remote-imds"
ARTIFACT_URL="https://public-artifact-bucket-382098889955-ap-northeast-1.s3.ap-northeast-1.amazonaws.com/aws-remote-imds/latest/amazonlinux2/arm64/ec2-remote-imds"

set -ex

# set arg and validation
MIDDLEWARE=$1
if [ "${MIDDLEWARE}" != "nginx" ] && [ "${MIDDLEWARE}" != "httpd" ]; then
    echo "select the middleware in front of this app as first positional argument: httpd or nginx"
    exit 1
fi


# download binary
sudo mkdir -p ${APP_DIR} 
sudo wget -P ${APP_DIR} ${ARTIFACT_URL}
sudo chmod +x ${APP_DIR}/${APP_NAME}

# create middleware conf file
if [ "${MIDDLEWARE}" == "nginx" ]; then
sudo mkdir -p /etc/nginx/conf.d
sudo cat <<EOF | sudo tee /etc/nginx/conf.d/${APP_NAME}.conf
server {
    listen       80;
    server_name  _;

    location  /imds {
        proxy_pass      http://127.0.0.1:9876/imds;
    }
}
EOF
elif [ "${MIDDLEWARE}" == "httpd" ]; then
sudo mkdir -p /etc/httpd/conf.d
sudo cat <<EOF | sudo tee /etc/httpd/conf.d/${APP_NAME}.conf
<VirtualHost *:80>
    ProxyPreserveHost On 
    ProxyPass /imds http://127.0.0.1:9876/imds
    ProxyPassReverse /imds http://127.0.0.1:9876/imds
</VirtualHost>
EOF
fi

# create service unit file
cat <<"EOF" | sudo tee /etc/systemd/system/${APP_NAME}.service
[Unit]
Description=ec2-remote-imds

[Service]
ExecStart=/opt/aws-remote-imds/ec2-remote-imds -f /opt/aws-remote-imds/ec2-remote-imds-config.yaml
Type=simple
ExecStop=/bin/kill -WINCH ${MAINPID}
Restart=always

[Install]
WantedBy=multi-user.target
EOF
# disable by default
sudo systemctl daemon-reload 
sudo systemctl disable ${APP_NAME}

# create default config file
cat <<"EOF" | sudo tee ${APP_DIR}/ec2-remote-imds-config.yaml
V1Url: ${IMDS_V1_URL|http://169.254.169.254}
V2Url: ${IMDS_V2_URL|http://169.254.169.254}

BasicAuth:
  Enabled: true
  Username: sample-user
  Password: sample-pass!!

AllowPathPrefixes:
  - /latest/api/token
  - /latest/meta-data/ami-id
  - /latest/meta-data/local-ipv4
EOF

