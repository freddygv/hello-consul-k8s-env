VERSION = v0.2.2
ACCOUNT = freddygv
APP = hello-http

all: push-docker

build:
	go build -o bin/hello

build-docker: build
	docker build -t $(ACCOUNT)/$(APP):$(VERSION) .

push-docker: build-docker
	docker push $(ACCOUNT)/$(APP):$(VERSION)