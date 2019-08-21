#!/bin/bash

# Loads a ConfigMap that allows pods to use the `.consul` TLD.
# Below we provide the config for both kube-dns and coredns.
# As of Kubernetes 1.13 CoreDNS is the default cluster DNS server.
#
# Call with the name of your DNS service as deployed by the Consul Helm chart.
#
#     configure-dns.sh piquant-shark-consul-dns
#
# https://www.consul.io/docs/platform/k8s/dns.html

if [ -n "$1" ]; then
  DNS_SERVICE_NAME=$1
else
  DNS_SERVICE_NAME="hedgehog-consul-dns"
fi

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health
        kubernetes cluster.local in-addr.arpa ip6.arpa {
            pods insecure
            upstream
            fallthrough in-addr.arpa ip6.arpa
            ttl 30
        }
        prometheus :9153
        proxy . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
    consul {
      errors
      cache 30
      proxy . $(kubectl get svc $DNS_SERVICE_NAME -o jsonpath='{.spec.clusterIP}')
    }
EOF


# Config if using kube-dns
#
#cat <<EOF | kubectl apply -f -
#apiVersion: v1
#kind: ConfigMap
#metadata:
#  labels:
#    addonmanager.kubernetes.io/mode: EnsureExists
#  name: kube-dns
#  namespace: kube-system
#data:
#  stubDomains: |
#    {"consul": ["$(kubectl get svc $DNS_SERVICE_NAME -o jsonpath='{.spec.clusterIP}')"]}
#EOF
