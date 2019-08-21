all: push-docker

push-docker:
	make -C hello-http/
	make -C hello-http-init/
	make -C hello-ttl/
	make -C hello-ttl-init/
	make -C hello-client/
	make -C hello-client-init/

deps:
	git clone https://github.com/hashicorp/consul-helm.git
	cd consul-helm && git checkout v0.8.1