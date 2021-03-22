# Consul 1.10 alpha - TransparentProxy MVP

### Background

Consul service mesh is a feature-set that was introduced to provide encrypted communication between services over mTLS. Inside of the service mesh services are deployed with sidecar proxies. These proxies enable functionality ranging from authorization and encryption to dynamic routing. 

For services to connect to each other through the mesh, each of them needs to dial its destinations through the local sidecar proxy. Proxies are configured by explicitly defining the set of “upstream” services that they need to connect to. This is done by listing the set of upstream services in the source service’s proxy registration. For each upstream service users also need to define a local port through which to dial that upstream service. 

To avoid explicitly defining upstreams and local ports to reach them on, the transparent proxy feature set automatically redirects inbound and outbound traffic through the sidecar proxies. Once received by the proxy, the traffic will be forwarded to the originally intended destination. Capturing and redirecting traffic reduces the chance that an application could circumvent the local proxy.

Traffic redirection in and out of the host allows applications to continue to address each other without modification. In the case of Kubernetes, applications can address each other using the native kube-dns, rather than hard coded proxy addresses such as `localhost:9090`.


### Functionality/limitations in the alpha:

- Kubernetes services can address each other using Pod IP or Cluster IP (including kube-dns).
- The Kubernetes service names MUST match the  corresponding Consul service names.
- For multi-cluster communication via mesh gateways, the upstream must still be defined explicitly via pod annotations or via proxy service registrations. Multi-cluster transparent proxying is not currently available.

### Prerequisites
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm with the HashiCorp Helm repository added](https://github.com/hashicorp/consul-helm)
- [Consul 1.10-alpha](https://releases.hashicorp.com/consul/) in your PATH


#### Transparent Proxy
Start a local Kubernetes kind cluster.

`kind create cluster --name tproxy`

Apply the deployments. This setup contains a simple client and server. 

The client dials the server using kube-dns, and the server replies with "Hello World" (You can monitor the status of the deployment with `watch kubectl get pods`).

`kubectl apply -f deployments/`

Once the containers are ready, check the client logs. You should see a recurring stream of "Hello World" responses from the server.

`kubectl logs -f -l "app=hello-client" -c hello-client`


Install the Consul using the helm chart with config from `helm-consul-values.yaml`.

`helm install -f helm-consul-values.yaml consul hashicorp/consul --version "0.31.0"`


Wait for the Consul deployent to finish and all the pods to be ready. Once that happens, port forward the Consul server so that commands from your terminal can be run against the Consul server in the Kubernetes cluster:

`kubectl port-forward consul-consul-server-0 8500:8500`


Submit a `proxy-defaults` config entry that enables "TransparentProxy" mode.
This setting will allow Consul to configure Envoy under the assumption that traffic is being redirected into the proxy. By setting the flag in a `proxy-defaults` config entry it applies to all services.

This flag is intended to be backwards compatible and should not break existing upstreams defined via Kubernetes pod annotations or proxy registrations on Consul.

`consul config write defaults.hcl`


Note that if you get a decode error after the previous command about `TransparentProxy` not being a recognized key, double check that the Consul v1.10 alpha is in your PATH. This error suggests that the binary running `consul config write` is outdated.

Next, the application deployments need to be patched. Patch the server first, this first patch will add an Envoy sidecar proxy via the `connect-inject` annotation. 

`kubectl patch deployment hello -p "$(cat ./patches/server-1.yaml)"`

Traffic from the client should only be briefly interrupted during the restart because the server can still be reached directly.

Next patch the client deployment. This patch will add an Envoy sidecar proxy as well as a container that will apply iptables rules to redirect all inbound and outbound traffic through Envoy.

`kubectl patch deployment hello-client -p "$(cat ./patches/client.yaml)"`

Traffic from the client to the server will now be flowing through the mesh. The client application's request are now being captured, routed through Envoy, and then sent to the server's proxy.

To prevent services from dialing the server directly, apply a second patch to the server deployment. This will apply the iptables rules that redirect inbound and outbound traffic through Envoy.

`kubectl patch deployment hello -p "$(cat ./patches/server-2.yaml)"`


Once again, after a brief interruption traffic should be flowing again:

`kubectl logs -f -l "app=hello-client" -c hello-client`


The reason why the connection from client to server is allowed is that the Consul deployment's intentions are currently set to allow all cross-service traffic. 

In a separate terminal, create some intentions to see how they affect the client logs. Adding an intention that denies all traffic will block requests from the client:

`consul intention create -deny "*" "*"`


Adding an intention that allows traffic between these two services will enable the requests to resume:

`consul intention create hello-client hello`

What's changing here is not just that the hello server will accept the connection. In transparent proxy mode upstreams are inferred from intentions, so the client proxy's configuration adds or removes the hello server's endpoints.

Test out making a request to a third party URL:

`kubectl exec deployment/hello-client -- curl https://example.com`

Currently we allow traffic to destinations outside of the mesh without the need for an egress proxy or gateway. In an upcoming release we will provide a solution for authorization of traffic leaving the mesh. 


#### Teardown

`kind delete cluster --name tproxy`