FROM golang:1.12.9 AS builder
WORKDIR /client
COPY . .
RUN go mod download

# https://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN CGO_ENABLED=0 go build -o client .

FROM alpine:3.10
COPY --from=builder /client/client /usr/bin/client

RUN apk add socat curl tcpdump
#RUN adduser -D client
#USER client

ENTRYPOINT ["client"]