GO_FILES := $(shell find . -name '*.go')
OUT_VARS_SVC := 'build/variables-svc'

build/executor: $(GO_FILES)
	mkdir -p build
	GOOS=linux go build -o build/executor ./executor/lambda

build/executor_lambda.zip: build/executor
	cd build && zip executor_lambda.zip ./executor

$(OUT_VARS_SVC): $(GO_FILES)
	mkdir -p build
	GOOS=linux go build -o $(OUT_VARS_SVC) ./variables/server

.PHONY: apply clean deploy generate init

apply:
	cd ./infra && terraform apply

clean:
	rm -rf ./build

deploy: build/executor_lambda.zip $(OUT_VARS_SVC) apply

dump-test:
	echo $(GO_FILES)

generate:
	go generate ./...

init:
	go get ./...
	go get golang.org/x/sys/unix
	cd ./infra && terraform init