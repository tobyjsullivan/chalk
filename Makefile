GO_FILES := $(shell find . -name '*.go')
IN_API_LAMBDA_SRC := ./api/lambda
IN_API_LOCAL_SRC := ./api/local
BUILD_DIR := build
OUT_RESOLVE_SVC := $(BUILD_DIR)/resolver-svc
OUT_VARS_SVC := $(BUILD_DIR)/variables-svc
API_BUILD_LOCAL_EXE := api-local
API_BUILD_LOCAL := $(BUILD_DIR)/$(API_BUILD_LOCAL_EXE)
API_BUILD_LAMBDA_EXE := api
API_BUILD_LAMBDA := $(BUILD_DIR)/$(API_BUILD_LAMBDA_EXE)
OUT_API_PACKAGE := $(BUILD_DIR)/executor_lambda.zip

$(API_BUILD_LAMBDA): $(GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux go build -o $(API_BUILD_LAMBDA) $(IN_API_LAMBDA_SRC)

$(API_BUILD_LOCAL): $(GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux go build -o $(API_BUILD_LOCAL) $(IN_API_LOCAL_SRC)

$(OUT_API_PACKAGE): $(API_BUILD)
	cd $(BUILD_DIR) && zip ../$(OUT_API_PACKAGE) $(API_BUILD_LAMBDA_EXE)

$(OUT_VARS_SVC): $(GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux go build -o $(OUT_VARS_SVC) ./variables/server

$(OUT_RESOLVE_SVC): $(GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux go build -o $(OUT_RESOLVE_SVC) ./resolver/server

.PHONY: apply clean deploy docker generate init services

apply:
	cd ./infra && terraform apply

clean:
	rm -rf $(BUILD_DIR)

deploy: $(OUT_API_PACKAGE) services apply

docker: services $(API_BUILD_LOCAL)
	docker build -f docker/Dockerfile.resolver-svc -t chalk-resolver-svc .
	docker build -f docker/Dockerfile.variables-svc -t chalk-variables-svc .
	docker build -f docker/Dockerfile.api -t chalk-api .

dump-test:
	echo $(GO_FILES)

generate:
	go generate ./...

init:
	go get ./...
	go get golang.org/x/sys/unix
	cd ./infra && terraform init

services: $(OUT_RESOLVE_SVC) $(OUT_VARS_SVC)