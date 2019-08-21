service {
  name = "hello"
  port = 8080
  checks = [
    {
      id = "hello-http"
      name = "HTTP API on port 8080"
      http = "https://localhost:8080/healthz"
      tls_skip_verify = true
      method = "GET"
      interval = "1s"
    }
  ]
}