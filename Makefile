IMAGE_NAME ?= puppernetes-terraform
IMAGE_TAG ?= latest
SSH_KEY_PATH ?= ~/.ssh/id_jenkins_p9s_nonprod

WORK_DIR := /work

.PHONY: clean build container ssh_agent ssh_gitlab

all: clean build container ssh_agent ssh_gitlab

build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

container_create:
	# create/start container if needed
	if [ ! -f .container_id ] || [ -z "$$(cat .container_id 2> /dev/null)" ] || ! docker inspect $$(cat .container_id 2> /dev/null) > /dev/null; then \
		docker create \
		--env SSH_AUTH_SOCK=/tmp/ssh-auth-sock \
		$(IMAGE_NAME):$(IMAGE_TAG) \
		sleep 3600 > .container_id; \
	fi; \

container: container_create
	$(eval CONTAINER_ID := $(shell cat .container_id 2> /dev/null))
	docker start $(CONTAINER_ID)

clean:
	test -e .container_id && { docker rm -f $(shell cat .container_id 2> /dev/null); rm -rf .container_id; }; true

ssh_agent: container
	docker exec $(CONTAINER_ID) bash -c "test -e \$${SSH_AUTH_SOCK} || ssh-agent -a \"\$${SSH_AUTH_SOCK}\""
	cat ${SSH_KEY_PATH} | docker exec -i $(CONTAINER_ID) ssh-add -

ssh_gitlab: container
	docker exec $(CONTAINER_ID) ssh git@gitlab.jetstack.net

terraform_sync: container
	docker cp Rakefile $(CONTAINER_ID):$(WORK_DIR)
	docker cp tfvars $(CONTAINER_ID):$(WORK_DIR)
	docker cp network $(CONTAINER_ID):$(WORK_DIR)
	docker cp tools $(CONTAINER_ID):$(WORK_DIR)
	docker cp vault $(CONTAINER_ID):$(WORK_DIR)
	docker cp kubernetes $(CONTAINER_ID):$(WORK_DIR)

terraform_plan: container
	docker exec $(CONTAINER_ID) bundle exec rake terraform:plan TERRAFORM_NAME=$(TERRAFORM_NAME) TERRAFORM_ENVIRONMENT=$(TERRAFORM_ENVIRONMENT) TERRAFORM_STACK=$(TERRAFORM_STACK) TERRAFORM_PLAN=/work/terraform.plan TERRAFORM_DESTROY=$(TERRAFORM_DESTROY)

terraform_apply: container
	docker exec $(CONTAINER_ID) bundle exec rake terraform:apply TERRAFORM_NAME=$(TERRAFORM_NAME) TERRAFORM_ENVIRONMENT=$(TERRAFORM_ENVIRONMENT) TERRAFORM_STACK=$(TERRAFORM_STACK) TERRAFORM_PLAN=/work/terraform.plan

terraform_validate: container
	docker exec $(CONTAINER_ID) bundle exec rake terraform:validate

terraform_fmt: container
	docker exec $(CONTAINER_ID) bundle exec rake terraform:fmt

vault_secrets: container
	docker exec $(CONTAINER_ID) bundle exec rake vault:secrets TERRAFORM_ENVIRONMENT=$(TERRAFORM_ENVIRONMENT)

vault_setup_k8s: container
	docker exec $(CONTAINER_ID) bundle exec rake vault:setup_k8s TERRAFORM_ENVIRONMENT=$(TERRAFORM_ENVIRONMENT) TERRAFORM_NAME=$(TERRAFORM_NAME)

vault_kubeconfig: container
	docker exec $(CONTAINER_ID) bundle exec rake vault:kubeconfig TERRAFORM_ENVIRONMENT=$(TERRAFORM_ENVIRONMENT) TERRAFORM_NAME=$(TERRAFORM_NAME)
	docker cp $(CONTAINER_ID):$(WORK_DIR)/kubeconfig-tunnel kubeconfig-tunnel

vault_initialize: container
	docker exec $(CONTAINER_ID) bundle exec rake vault:initialize TERRAFORM_ENVIRONMENT=$(TERRAFORM_ENVIRONMENT)

puppet_deploy_env: ssh_agent
	docker cp puppet.tar.gz $(CONTAINER_ID):$(WORK_DIR)
	docker exec $(CONTAINER_ID) bundle exec rake puppet:deploy_env TERRAFORM_ENVIRONMENT=$(TERRAFORM_ENVIRONMENT) TERRAFORM_NAME=$(TERRAFORM_NAME)
