# Consul Connect on Kubernetes

### Prerequisites
- minikube
- kubectl
- helm

### Instructions
#### Setup
* Start minikube with 4GB to 8GB of memory.

`$ minikube start --memory 8192`

* Initialize helm on the minikube cluster.

`$ helm init`

* Clone `consul-helm` to the current directory and checkout the latest release tag.

`$ make deps`

* Open the Kubernetes dashboard.

`$ minikube dashboard`

* Install the `consul-helm` chart with the config in `helm-consul-values.yaml`.

`$ helm install -f helm-consul-values.yaml --name hedgehog ./consul-helm`

Note that if you see the error: `Error: could not find a ready tiller pod`, helm has not finished initializing.

* Deploy all applications to k8s:

`$ kubectl create -f deployments/`

* Open Consul's web UI.

`$ minikube service hedgehog-consul-ui`

#### Teardown
`minikube delete`