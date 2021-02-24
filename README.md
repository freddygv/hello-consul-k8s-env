# Kubernetes TransparentProxy MVP

### Prerequisites
- kind
- kubectl
- helm

### Instructions

#### Building custom Consul binary

* If you need to load a custom Consul build, run these from the Consul repo:

`make dev-docker`

`docker tag $(docker images | rg consul-dev | rg latest  | awk '{print $3}') freddygv/consul-dev:tproxy2-v0.06`

#### Setup
* Start a kind cluster.

`kind create cluster --name consul-k8s`

* IFF you built a custom Consul binary, load it into the kind cluster:

`kind load --name consul-k8s docker-image freddygv/consul-dev:tproxy2-v0.06`

* If needed, update the helm chart's docker image to match the newly loaded one.

* Apply the deployments

`kubectl apply -f deployments/`

* Check logs for client. You should see it working after the containers are running.

`kubectl logs -f -l "app=hello-client" -c hello-client`

* Install the consul helm chart with the config in `helm-consul-values.yaml`.

`helm install -f helm-consul-values.yaml consul hashicorp/consul --version "0.26.0"`

* Wait for the Consul deployent to finish. Once the consul pods/containers are running you need to patch the hello server. This patch adds an init container that will apply the iptables rules to capture inbound and outbound traffic:

`kubectl patch deployment hello -p "$(cat ./patches/server.yaml)"`

* Then patch the client deployment. This patch applies the same init container with iptables rules:

`kubectl patch deployment hello-client -p "$(cat ./patches/client.yaml)"`

* Check logs for client, should be failing after the patched pod starts. The new pod will have 3 containers.

`kubectl logs -f -l "app=hello-client" -c hello-client`

* Port forward Consul so that commands from your terminal can be run against the Consul server in the kind cluster:

`kubectl port-forward consul-consul-server-0 8500:8500`

* Write out proxy-default for TransparentProxy. This will allow Consul to configure Envoy based on intentions, rather than explicit upstreams.

`consul config write defaults.hcl`

* Create an intention between it and the server. This will enable Envoy to forward traffic to the hello server.

`consul intention create client hello`

* Check logs for client again, should see it working:

`kubectl logs -f -l "app=hello-client" -c hello-client`


* Delete the intention between the client and server

`consul intention delete client hello`

* Check logs for client, should fail, but it isn't at the moment.

`kubectl logs -f -l "app=hello-client" -c hello-client`


#### Debugging Envoy

* Review Envoy config for client by forwarding the Envoy admin API port:

`kubectl port-forward deployment/hello-client 19000:19000`

Then navigate to: http://localhost:19000/config_dump

* Check Envoy logs

`kubectl logs -f -l "app=hello-client" -c consul-connect-envoy-sidecar`

#### Teardown

`kubectl delete -f deployments/`

`helm del consul`

`kubectl delete pvc -l release=consul`

`kubectl get secret | grep consul | grep Opaque | grep token | awk '{print $1}' | xargs kubectl delete secret`

`kind delete cluster --name consul-k8s`