BUILD_IMAGE := fluentbit-repo:build
BUCKET_NAME := releases.tarmak.io

VERSION := unknown

image:
	docker build -t $(BUILD_IMAGE) --no-cache .

upload: image
	rm -rf ./output
	docker create $(BUILD_IMAGE) | xargs -n 1 -i docker cp {}:/output ./output
	gsutil rsync -d -r ./output/ gs://$(BUCKET_NAME)/fluentbit/$(VERSION)/centos-7/
