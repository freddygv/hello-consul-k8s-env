#!/bin/sh

cat <<EOF > hello-service.json
{
  "Name": "hello",
  "ID": "hello-http",
  "Address": "${POD_IP}",
  "Port": 8080,
  "Check": {
    "CheckID": "hello-http",
    "Name": "HTTP API on port 8080",
    "Method": "GET",
    "HTTP": "http://${POD_IP}:8080/healthz",
    "TLSSkipVerify": true,
    "Interval": "1s"
  }
}
EOF

curl -X PUT \
    --data @hello-service.json \
    "http://${HOST_IP}:8500/v1/agent/service/register"