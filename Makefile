GO_FILES := $(shell find . -name '*.go')
IN_API_SRC := ./api/lambda
BUILD_DIR := build
OUT_VARS_SVC := $(BUILD_DIR)/variables-svc
API_BUILD_EXE := api
API_BUILD := $(BUILD_DIR)/$(API_BUILD_EXE)
API_PACKAGE := $(BUILD_DIR)/executor_lambda.zip

$(API_BUILD): $(GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux go build -o $(API_BUILD) $(IN_API_SRC)

$(API_PACKAGE): $(API_BUILD)
	cd $(BUILD_DIR) && zip ../$(API_PACKAGE) $(API_BUILD_EXE)

$(OUT_VARS_SVC): $(GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux go build -o $(OUT_VARS_SVC) ./variables/server

.PHONY: apply clean deploy generate init

apply:
	cd ./infra && terraform apply

clean:
	rm -rf $(BUILD_DIR)

deploy: $(API_PACKAGE) $(OUT_VARS_SVC) apply

dump-test:
	echo $(GO_FILES)

generate:
	go generate ./...

init:
	go get ./...
	go get golang.org/x/sys/unix
	cd ./infra && terraform init