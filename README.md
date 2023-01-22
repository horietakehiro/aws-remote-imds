# aws-remote-imds

aws-remote-imds is a reverse proxy server that is supposed to be running on AWS EC2 instances or ECS Fargate containers. aws-remote-imds routes your web requests to the [imds](https://docs.aws.amazon.com/ja_jp/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) endpoint - http://169.254.169.254, and return the metadata with some additional information.

You can inspect ec2 instances or ecs fargate containers' metadata via http api, without directry connecting them via ssh or session manager.

---

## Detail

- ec2-remote-imds: See [here](./doc/ec2/README.md)
- fargate-remote-imds: **Not Implemented Yet**


---