EXECUTABLE := auto-rpm

LDFLAGS ?= -X 'main.Version=1.0'

# for dockerhub
DEPLOY_ACCOUNT := cody0704
DEPLOY_IMAGE:= $(EXECUTABLE)
TAGS ?=


build_linux_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w $(LDFLAGS)' -o release/linux/amd64/$(DEPLOY_IMAGE)

build_linux_i386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -ldflags '-w $(LDFLAGS)' -o release/linux/i386/$(DEPLOY_IMAGE)

build_linux_arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags '-w $(LDFLAGS)' -o release/linux/arm64/$(DEPLOY_IMAGE) 

docker: docker_image

docker_image:
	docker build -t $(DEPLOY_ACCOUNT)/$(DEPLOY_IMAGE) .

docker_deploy:
ifeq ($(tag),)
	@echo "Usage: make $@ tag=<tag>"
	@exit 1
endif
	# deploy image
	docker tag $(DEPLOY_ACCOUNT)/$(DEPLOY_IMAGE):latest $(DEPLOY_ACCOUNT)/$(DEPLOY_IMAGE):$(tag)
	docker push $(DEPLOY_ACCOUNT)/$(DEPLOY_IMAGE):latest
	docker push $(DEPLOY_ACCOUNT)/$(DEPLOY_IMAGE):$(tag)