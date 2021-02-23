# Kubernetes TransparentProxy MVP

### Prerequisites
- kind
- kubectl
- helm

### Instructions

#### Setup
* Start a kind cluster.

`kind create cluster --name consul-k8s`

* If you need to load a new custom Consul build, run these from the Consul repo:

`make dev-docker`

`docker tag $(docker images | rg consul-dev | rg latest  | awk '{print $3}') freddygv/consul-dev:tproxy2-v0.06`

* Load customer image into kind cluster

`kind load --name consul-k8s docker-image freddygv/consul-dev:tproxy2-v0.06`

* If needed, update the helm chart's docker image to this one

* Apply the deployments

`kubectl apply -f deployments/`

* Check logs for client, should see it working

`kubectl logs -f -l "app=hello-client" -c hello-client`

* Install the consul helm chart with the config in `helm-consul-values.yaml`.

`helm install -f helm-consul-values.yaml consul hashicorp/consul --version "0.26.0"`

* Patch the server:

`kubectl patch deployment hello -p "$(cat ./patches/server.yaml)"`

* Patch client deployment

`kubectl patch deployment hello-client -p "$(cat ./patches/client.yaml)"`

* Check logs for client, should be failing after restarting

`kubectl logs -f -l "app=hello-client" -c hello-client`

* Port forward Consul

`kubectl port-forward consul-consul-server-0 8500:8500`

* Write out proxy-default for TransparentProxy

`consul config write defaults.hcl`

* Create an intention between it and the server

`consul intention create client hello`

* Check logs for client again, should see it working

`kubectl logs -f -l "app=hello-client" -c hello-client`

* Review Envoy config for client

`kubectl port-forward deployment/hello-client 19000:19000`

* Delete the intention between the client and server

`consul intention delete client hello`

* Check logs for client, should be failing again

`kubectl logs -f -l "app=hello-client" -c hello-client`

* Check Envoy logs

`kubectl logs -f -l "app=hello-client" -c consul-connect-envoy-sidecar`

#### Teardown
`kubectl delete -f deployments/`

`helm del consul`

`kubectl delete pvc -l release=consul`

`kubectl get secret | grep consul | grep Opaque | grep token | awk '{print $1}' | xargs kubectl delete secret`

`kind delete cluster --name consul-k8s`