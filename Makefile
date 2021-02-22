all: push-docker

push-docker:
	make -C hello-http/
	make -C hello-client/