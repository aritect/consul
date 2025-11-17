REGISTRY := docker-hosted.nexus.infrastructure.alxshelepenok.com
CONSUL_IMAGE := aritect/consul
TAG := latest

DOCKER_BUILD = docker build
DOCKER_TAG = docker tag
DOCKER_PUSH = docker push
DOCKER_RMI = docker rmi -f

define check_nexus_vars
	@if [ -z "$(NEXUS_USERNAME)" ]; then \
		echo "ERROR: NEXUS_USERNAME is not set"; \
		exit 1; \
	fi
	@if [ -z "$(NEXUS_PASSWORD)" ]; then \
		echo "ERROR: NEXUS_PASSWORD is not set"; \
		exit 1; \
	fi
endef

define build_service
	@echo "Building $(1) service..."
	$(call check_nexus_vars)
	$(DOCKER_BUILD) -f Dockerfile -t $(2):$(TAG) $(NEXUS_BUILD_ARGS) $(3) .
endef

define push_service
	@echo "Pushing $(1) service..."
	$(DOCKER_TAG) $(2):$(TAG) $(REGISTRY)/$(2):$(TAG)
	$(DOCKER_PUSH) $(REGISTRY)/$(2):$(TAG)
endef

.PHONY: docker-login
docker-login:
	@echo "Logging in to $(REGISTRY)..."
	$(call check_nexus_vars)
	echo "$(NEXUS_PASSWORD)" | docker login $(REGISTRY) -u "$(NEXUS_USERNAME)" --password-stdin

.PHONY: build
build:
	$(call build_service,consul,$(CONSUL_IMAGE))

.PHONY: push
push: build
	$(call push_service,consul,$(CONSUL_IMAGE))

.PHONY: clean
clean:
	$(DOCKER_RMI) $(CONSUL_IMAGE):$(TAG)
