VERSION = v0.1.0
ACCOUNT = freddygv
APP = hello-http-init

all: push-docker

build-docker:
	docker build -t $(ACCOUNT)/$(APP):$(VERSION) .

push-docker: build-docker
	docker push $(ACCOUNT)/$(APP):$(VERSION)