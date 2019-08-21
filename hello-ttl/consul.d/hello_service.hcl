service {
  name = "hello"
  port = 8080
  checks = [
    {
      id = "hello-ttl"
      name = "5s TTL"
      ttl = "5s"
    }
  ]
}