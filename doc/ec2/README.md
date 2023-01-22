# ec2-remote-imds

---

## Example of Usage

- Launch Server

```Bash
$ ./ec2-remote-imds -f path/to/config.yaml
2023/01/22 10:53:31 use http://169.254.169.254 as imds v1 url
2023/01/22 10:53:31 use http://169.254.169.254 as imds v2 url
   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v4.10.0
High performance, minimalist Go web framework
https://echo.labstack.com
____________________________________O/_______
                                    O\
⇨ http server started on [::]:9876
```

- Invole Endpoint

You can invoke endpoint with `/imds/v1/*` for imds v1, and `/imds/v2/*` for imds v2.

```Bash
# for v1
$ curl -u "test-user:test-pass" \
    http://127.0.0.1:9876/imds/v1/latest/ \
    | jq
{
  "InstanceMetadata": {
    "QueryPath": "/latest/",
    "Value": null,
    "Options": [
      "dynamic",
      "meta-data",
      "user-data"
    ],
    "Error": null
  },
  "RequestMetadata": {
    "Proto": "HTTP/1.1",
    "X-Forwarded-For": [
      "127.0.0.1"
    ],
    "X-Real-Ip": [
      "127.0.0.1"
    ],
    "X-Forwarded-Proto": [
      "http"
    ]
  },
  "ResponseMetadata": {}
}

$ curl -u "test-user:test-pass" \
    http://127.0.0.1:9876/imds/v1/latest/meta-data/ami-id \
    | jq .InstanceMetadata
{
  "InstanceMetadata": {
    "QueryPath": "/latest/meta-data/ami-id/",
    "Value": "ami-0a887e401f7654935",
    "Options": [],
    "Error": null
  }
}

# for v2
$ TOKEN=`curl -X PUT \
    -u "test-user:test-pass" \
    -H "X-aws-ec2-metadata-token-ttl-seconds: 21600" \
    http://127.0.0.1:9876/imds/v2/latest/api/token \
    | jq -r .InstanceMetadata.Value`

$ curl -u "test-user:test-pass" \
    -H "X-aws-ec2-metadata-token: $TOKEN" \
    http://127.0.0.1:9876/imds/v2/latest/meta-data/ami-id \
    | jq .InstanceMetadata

{
  "InstanceMetadata": {
    "QueryPath": "/latest/meta-data/ami-id/",
    "Value": "ami-0a887e401f7654935",
    "Options": [],
    "Error": null
  }
}
```

---


## Features

- Basic Auth
    - For protecting the endpoint, ec2-remote-imds enables basic auth by default. You can disable it, or customize username and password by editing [configuration file](../../config/ec2/ec2-remote-imds-config.yaml).
- Path Restriction
    - For reducing security risk, you can restrict the list of accessible request url path.
    - For example, if you set `/latest/meta-data/` in your configuration file's `AllowPathPrefixes` section, all metadata requests whose url path starts with `/latest/meta-data/` will be accessible but all others like `/latest/user-data/` will be denyed.

- Token for v2 as Response Cookie and Header
    - When you request the path `/imds/v2/latest/api/token` to retreive `X-aws-ec2-metadata-token`, it is also responsed as cookie and header like below.
```Bash
$ curl --head -X PUT \
    -u "test-user:test-pass" \
    -H "X-aws-ec2-metadata-token-ttl-seconds: 21600" \
    http://127.0.0.1:9876/imds/v2/latest/api/token
HTTP/1.1 200 OK
Content-Length: 284
Content-Type: application/json
Date: Sun, 22 Jan 2023 02:15:36 GMT
Set-Cookie: X-aws-ec2-metadata-token=THIVo7U56x5YScYHfbtXIvVxeiiaJm+XZHmBmY6+qJw=; Expires=Sun, 22 Jan 2023 08:15:36 GMT
X-Aws-Ec2-Metadata-Token: THIVo7U56x5YScYHfbtXIvVxeiiaJm+XZHmBmY6+qJw=
X-Aws-Ec2-Metadata-Token-Ttl-Seconds: 21600
```

---

## Install

- Install command snippet:

```Bash
$ OS_NAME="amazonlinux2" # or ubuntu
$ ARCH_NAME="amd64" # currently, arm64 is not supported(but will be soon supported)
$ MIDDLEWARE="nginx" # or httpd
$ wget -O - https://public-artifact-bucket-382098889955-ap-northeast-1.s3.ap-northeast-1.amazonaws.com/aws-remote-imds/latest/${OS_NAME}/${ARCH_NAME}/install.sh | sudo bash -s ${MIDDLEWARE}
```

- What has been created:

```Bash
# middleware configuration file has been created
# if you select nginx as middleware:
$ cat /etc/nginx/conf.d/ec2-remote-imds.conf
server {
    listen       80;
    server_name  _;

    location  /imds {
        proxy_pass      http://127.0.0.1:9876/imds;
    }
}
# if you select httpd as middleware:
$ cat /etc/httpd/conf.d/ec2-remote-imds.conf
# in ubuntu, `$ cat /etc/apache2/sites-enabled/ec2-remote-imds.conf`
<VirtualHost *:80>
    ProxyPreserveHost On
    ProxyPass /imds http://127.0.0.1:9876/imds
    ProxyPassReverse /imds http://127.0.0.1:9876/imds
</VirtualHost>

# systemd unit file has been created
$ cat /etc/systemd/system/ec2-remote-imds.service
[Unit]
Description=ec2-remote-imds

[Service]
ExecStart=/opt/aws-remote-imds/ec2-remote-imds -f /opt/aws-remote-imds/ec2-remote-imds-config.yaml
Type=simple
ExecStop=/bin/kill -WINCH ${MAINPID}
Restart=always

[Install]
WantedBy=multi-user.target

# by default, ec2-remote-imds is stopped and disabled
$ systemctl is-enabled ec2-remote-imds.service
disabled
$ systemctl status ec2-remote-imds.service
systemctl status ec2-remote-imds.service
● ec2-remote-imds.service - ec2-remote-imds
   Loaded: loaded (/etc/systemd/system/ec2-remote-imds.service; disabled; vendor preset: disabled)
   Active: inactive (dead)
```

**After install and before launch, You should replace config file with you own one - customize allow path prefix and username, password.**

---

## Configuration

For the example and full reference for configuration file, see [here](../../config/ec2/ec2-remote-imds-config.yaml)

---