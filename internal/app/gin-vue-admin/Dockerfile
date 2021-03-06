
#FROM alpine:latest
FROM debian:stretch

WORKDIR /app

#ADD ./gva ./
ADD ./lianmi-admin-server-linux-amd64 ./
ADD ./config.yaml ./
ADD ./resource ./resource

EXPOSE 8888

CMD ["./lianmi-admin-server-linux-amd64"]

