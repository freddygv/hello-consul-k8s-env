VERSION = v0.1.0
ACCOUNT = freddygv
APP = hello-client

all: push-docker

build:
	go build -o bin/client

build-docker:
	docker build -t $(ACCOUNT)/$(APP):$(VERSION) .

push-docker: build-docker
	docker push $(ACCOUNT)/$(APP):$(VERSION)