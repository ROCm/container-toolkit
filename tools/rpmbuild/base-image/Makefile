all:
	@if [ -n "$(INSECURE_REGISTRY)" ]; then \
	echo "building rpmbuild dev container with insecure registry" ;\
	docker build --build-arg USER=$(shell id -un) --build-arg GROUP=$(shell id -gn) \
		--build-arg BUILD_BASE_IMAGE=$(BUILD_BASE_IMAGE) \
		--build-arg INSECURE_REGISTRY=$(INSECURE_REGISTRY) \
		--build-arg UID=$(shell id -u) --build-arg GID=$(shell id -g) -t ${RPM_BUILD_CONTAINER} . ;\
	else \
	echo "building rpmbuild dev container without insecure registry" ;\
	docker build --build-arg USER=$(shell id -un) --build-arg GROUP=$(shell id -gn) \
		--build-arg BUILD_BASE_IMAGE=$(BUILD_BASE_IMAGE) \
		--build-arg UID=$(shell id -u) --build-arg GID=$(shell id -g) -t ${RPM_BUILD_CONTAINER} . ;\
	fi
	@echo "rpmbuild dev container build complete : ${RPM_BUILD_CONTAINER}"
