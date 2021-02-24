VERSION = v0.2.4
ACCOUNT = freddygv
APP = hello-client

all: push-docker

build:
	go build -o bin/client

build-docker: build
	docker build -t $(ACCOUNT)/$(APP):$(VERSION) .

push-docker: build-docker
	docker push $(ACCOUNT)/$(APP):$(VERSION)