.PHONY: init deploy clean

init:
	go get -t -v ./...

deploy: build/handler.zip
	terraform apply

build/handler.zip: $(wildcard **/*.go)
	mkdir -p build
	GOOS=linux go build -o build/handler ./lambda
	cd build && zip handler.zip ./handler

clean:
	rm -rf ./build