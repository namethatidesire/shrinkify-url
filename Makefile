.PHONY: test build deploy

TABLE_NAME ?= urls
QUEUE_URL ?= ""
FUNCTION_NAME ?= shrinkify-url

test:
	go test ./...

build:
	GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/lambda
ifeq ($(OS),Windows_NT)
	powershell Compress-Archive -Force -Path bootstrap -DestinationPath function.zip
else
	zip function.zip bootstrap
endif

deploy: build
	aws lambda update-function-code \
	--function-name $(FUNCTION_NAME) \
	--zip-file fileb://function.zip