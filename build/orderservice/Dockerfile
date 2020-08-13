FROM alpine:latest

RUN apk update && apk add bash

ADD ./dist/orderservice-linux-amd64 app
ADD ./configs/orderservice.yml .
ADD ./scripts/wait-for .


