GO_FILES := $(shell find . -name '*.go')
PROTO_FILES := $(shell find . -name '*.proto')
TF_FILES := $(shell find . -name '*.tf')
IN_API_LAMBDA_SRC := ./api/lambda
IN_API_LOCAL_SRC := ./api/local
BUILD_DIR := build
OUT_RESOLVE_SVC := $(BUILD_DIR)/resolver-svc
OUT_MONOLITH_SVC := $(BUILD_DIR)/monolith-svc
API_BUILD_LOCAL_EXE := api-local
API_BUILD_LOCAL := $(BUILD_DIR)/$(API_BUILD_LOCAL_EXE)
API_BUILD_LAMBDA_EXE := api
API_BUILD_LAMBDA := $(BUILD_DIR)/$(API_BUILD_LAMBDA_EXE)
OUT_API_PACKAGE := $(BUILD_DIR)/executor_lambda.zip
DOCKER_IMAGES := ./docker/images

$(API_BUILD_LAMBDA): precompile $(GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux go build -o $(API_BUILD_LAMBDA) $(IN_API_LAMBDA_SRC)

$(API_BUILD_LOCAL): precompile $(GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux go build -o $(API_BUILD_LOCAL) $(IN_API_LOCAL_SRC)

$(OUT_API_PACKAGE): $(API_BUILD_LAMBDA)
	cd $(BUILD_DIR) && zip ../$(OUT_API_PACKAGE) $(API_BUILD_LAMBDA_EXE)

$(OUT_MONOLITH_SVC): precompile $(GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux go build -o $(OUT_MONOLITH_SVC) ./monolith/server

$(OUT_RESOLVE_SVC): precompile $(GO_FILES)
	mkdir -p $(BUILD_DIR)
	GOOS=linux go build -o $(OUT_RESOLVE_SVC) ./resolver/server

$(DOCKER_IMAGES)/api.tar.gz: docker/Dockerfile.api $(API_BUILD_LOCAL)
	mkdir -p $(DOCKER_IMAGES)
	docker build -f docker/Dockerfile.api -t chalk-api .
	docker save chalk-api:latest > $(DOCKER_IMAGES)/api.tar.gz

$(DOCKER_IMAGES)/resolver-svc.tar.gz: docker/Dockerfile.resolver-svc $(OUT_RESOLVE_SVC)
	mkdir -p $(DOCKER_IMAGES)
	docker build -f docker/Dockerfile.resolver-svc -t chalk-resolver-svc .
	docker save chalk-resolver-svc:latest > $(DOCKER_IMAGES)/resolver-svc.tar.gz

$(DOCKER_IMAGES)/monolith-svc.tar.gz: docker/Dockerfile.monolith-svc $(OUT_MONOLITH_SVC)
	mkdir -p $(DOCKER_IMAGES)
	docker build -f docker/Dockerfile.monolith-svc -t chalk-monolith-svc .
	docker save chalk-monolith-svc:latest > $(DOCKER_IMAGES)/monolith-svc.tar.gz

.PHONY: apply clean compose deploy docker format generate init push-docker precompile test

apply:
	cd ./infra && terraform apply

clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(DOCKER_IMAGES)

compose: docker-compose.yml docker
	docker-compose up

deploy: $(OUT_API_PACKAGE) $(OUT_RESOLVE_SVC) $(OUT_MONOLITH_SVC) apply push-docker

docker: $(DOCKER_IMAGES)/resolver-svc.tar.gz $(DOCKER_IMAGES)/monolith-svc.tar.gz $(DOCKER_IMAGES)/api.tar.gz
	docker load < $(DOCKER_IMAGES)/resolver-svc.tar.gz
	docker load < $(DOCKER_IMAGES)/monolith-svc.tar.gz
	docker load < $(DOCKER_IMAGES)/api.tar.gz

dump-test:
	echo $(GO_FILES)

format: $(GO_FILES) $(TF_FILES)
	go fmt ./...
	goimports -w ./
	cd ./infra && terraform format

generate: $(PROTO_FILES) $(GO_FILES)
	go generate ./...

init:
	go get ./...
	go get golang.org/x/sys/unix
	cd ./infra && terraform init

push-docker: docker
	$$(aws ecr get-login --region $$(cd ./infra && terraform output aws_region) --no-include-email)
	docker tag chalk-monolith-svc "$$(cd ./infra && terraform output repo_monolith_svc_url):latest"
	docker push "$$(cd ./infra && terraform output repo_monolith_svc_url):latest"
	docker tag chalk-resolver-svc "$$(cd ./infra && terraform output repo_resolver_svc_url):latest"
	docker push "$$(cd ./infra && terraform output repo_resolver_svc_url):latest"
	docker tag chalk-api "$$(cd ./infra && terraform output repo_api_url):latest"
	docker push "$$(cd ./infra && terraform output repo_api_url):latest"

precompile: format generate

test:
	go test ./...