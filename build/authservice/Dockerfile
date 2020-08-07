FROM alpine:latest

RUN apk update && apk add bash

ADD ./dist/authservice-linux-amd64 app
ADD ./configs/authservice.yml .
ADD ./scripts/wait-for .


