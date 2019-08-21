FROM alpine:3.10
RUN apk add --no-cache curl
RUN adduser -D init
USER init
WORKDIR /home/init
COPY . /home/init
CMD ["./register-consul-service.sh"]