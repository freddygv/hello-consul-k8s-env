# Consul Connect on Kubernetes

### Prerequisites
- minikube
- kubectl
- helm

### Instructions
#### Setup
* Start minikube with 4GB to 8GB of memory.

```$ minikube start --memory 8192```

* Initialize helm on the minikube cluster.

```$ helm init```

* Clone `consul-helm` to the current directory and checkout the latest release tag.

```$ make deps```

* Open the Kubernetes dashboard.

```$ minikube dashboard```

* Install the `consul-helm` chart with the config in `helm-consul-values.yaml`.

```$ helm install -f helm-consul-values.yaml --name hedgehog ./consul-helm```

Note that if you see the error: `Error: could not find a ready tiller pod`, helm has not finished initializing.

* Deploy all applications to k8s:

```$ kubectl create -f deployments/```

* Open Consul's web UI.

```$ minikube service hedgehog-consul-ui```

##### Dynamic Configuration
* Get the list of pods and find one that is running a Consul agent. 
We'll use this as an easy way to run Consul CLI commands.

```$ kubectl get pods```

* Look for one with consul in the name and connect to the running pod.

`$ kubectl exec -it hedgehog-consul-5t2dc /bin/sh`

* Once connected, run a command that saves a value to Consul.

`$ consul kv put service/hello/hello-http/enable_checks false`

* Switch to the Consul UI and note that the HTTP check for the hello service is failing.

#### Teardown
`minikube delete`