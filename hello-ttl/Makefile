VERSION = v0.1.0
ACCOUNT = freddygv
APP = hello-ttl

all: push-docker

build:
	go build -o bin/hello

build-docker:
	docker build -t $(ACCOUNT)/$(APP):$(VERSION) .

push-docker: build-docker
	docker push $(ACCOUNT)/$(APP):$(VERSION)