ec2-run-mock: ec2-stop-mock
	docker-compose -f tool/ec2/docker-compose.yaml up -d
	docker-compose -f tool/ec2/docker-compose.yaml ps
ec2-stop-mock:
	docker-compose -f tool/ec2/docker-compose.yaml down

ec2-local-run: export IMDS_V1_URL=http://localhost:1111
ec2-local-run: export IMDS_V2_URL=http://localhost:2222
ec2-local-run:
	go run ./cmd/ec2/main.go -f config/ec2/ec2-remote-imds-config.yaml


ec2-run:
	go run ./cmd/ec2/main.go

ec2-build:
	go build -o ./bin/ec2-remote-imds ./cmd/ec2/

ec2-test: export IMDS_V1_URL=http://localhost:1111
ec2-test: export IMDS_V2_URL=http://localhost:2222
ec2-test:
	go test ./...

ec2-cicd-deploy:
	cd deployments/cdk/ && cdk deploy --stack Ec2CiCdStach --require-approval=never --no-rollback