.PHONY: deploy

deploy:
	rm -rf ./build
	yarn build
	aws s3 sync ./build/ s3://messy.codes/

build:
	parcel build ./src/index.tsx

serve:
	parcel ./src/index.html