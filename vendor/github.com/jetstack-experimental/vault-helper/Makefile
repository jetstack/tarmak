REGISTRY := jetstackexperimental
IMAGE_NAME := vault-helper
IMAGE_TAGS := canary
BUILD_TAG := build

image:
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(BUILD_TAG) .

push: image
	set -e; \
	for tag in $(IMAGE_TAGS); do \
		docker tag $(REGISTRY)/$(IMAGE_NAME):$(BUILD_TAG) $(REGISTRY)/$(IMAGE_NAME):$${tag} ; \
		docker push $(REGISTRY)/$(IMAGE_NAME):$${tag}; \
	done

test:
	docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(CURDIR):/code --workdir /code ruby:2.3 bash -c "bundle install && bundle exec rake"


go_codegen:
	mockgen -package kubernetes -source=pkg/kubernetes/kubernetes.go > pkg/kubernetes/kubernetes_mocks_test.go
