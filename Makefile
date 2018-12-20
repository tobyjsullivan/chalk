.PHONY: init deploy clean

deploy: build/handler.zip
	terraform apply

build/handler.zip: $(wildcard **/*.go)
	mkdir -p build
	GOOS=linux go build -o build/executor ./executor
	cd build && zip executor_lambda.zip ./executor

clean:
	rm -rf ./build