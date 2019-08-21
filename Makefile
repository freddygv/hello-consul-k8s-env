all: push-docker

push-docker:
	make -C hello-http/
	make -C hello-client/

deps:
	git clone https://github.com/hashicorp/consul-helm.git
	cd consul-helm && git checkout v0.8.1