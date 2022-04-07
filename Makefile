BUCKET_NAME := lacework-alliances
KEY_PREFIX := lacework-aws-security-hub
LAMBDA_PREFIX := lambda/
CFT_PREFIX := templates/
HONEY_DATASET := lacework-alliances-prod

PROFILE ?= ct
REGION ?= us-west-2

all: build

.PHONY: clean build

clean:
	rm main || true
	rm main_new.go || true
	rm setup.zip || true

build: clean
	buildid=$$(git describe --all --long | cut -d "/" -f 2); \
	sed -e "s|\$$BUILD|$$buildid|g" -e "s|\$$DATASET|$(HONEY_DATASET)|g" -e "s|\$$HONEY_KEY|$(HONEY_KEY)|g" main.go > main_new.go; \
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o main main_new.go
	zip setup.zip main
	@aws --region $(REGION) s3 cp setup.zip s3://$(BUCKET_NAME)/$(KEY_PREFIX)/$(LAMBDA_PREFIX) --acl public-read
	@aws --region $(REGION) s3 cp aws-security-hub-integration.yml s3://$(BUCKET_NAME)/$(KEY_PREFIX)/$(CFT_PREFIX) --acl public-read
	rm main || true
	rm main_new.go || true
	rm setup.zip || true






