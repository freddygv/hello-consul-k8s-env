FROM golang:1.12.9 AS builder
WORKDIR /hello
COPY . .
RUN go mod download

# https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN CGO_ENABLED=0 go build -o hello .

FROM alpine:3.10
COPY --from=builder /hello/hello /usr/bin/hello

RUN apk add socat curl tcpdump
RUN adduser -D hello
USER hello
ENTRYPOINT ["hello"]