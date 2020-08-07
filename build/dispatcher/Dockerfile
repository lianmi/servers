FROM alpine:latest

RUN apk update && apk add bash

ADD ./dist/dispatcher-linux-amd64 app
ADD ./configs/dispatcher.yml .
ADD ./scripts/wait-for .


