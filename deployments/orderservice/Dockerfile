# FROM alpine:latest
# FROM frolvlad/alpine-glibc:latest
FROM debian:stretch

# wait-for-it.sh:
# curl -O https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh
# chmod +x wait-for-it.sh

ADD ./orderservice-linux-amd64 /orderservice
ADD ./orderservice.yml /
ADD ./wait-for-it.sh /

EXPOSE 28080

# ENTRYPOINT ["/bin/bash", "-c"]
CMD ["/wait-for-it.sh" , "db:3306" , "--strict" , "--timeout=300" , "--" , \
  "/orderservice", "-f", "/orderservice.yml"]
