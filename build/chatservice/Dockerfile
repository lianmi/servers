FROM alpine:latest

RUN apk update && apk add bash

ADD ./dist/chatservice-linux-amd64 app
ADD ./configs/chatservice.yml .
ADD ./scripts/wait-for .


