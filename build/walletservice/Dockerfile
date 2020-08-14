FROM alpine:latest

RUN apk update && apk add bash

ADD ./dist/walletservice-linux-amd64 app
ADD ./configs/walletservice.yml .
ADD ./scripts/wait-for .


