# FROM alpine:latest
# FROM frolvlad/alpine-glibc:latest
FROM debian:stretch

# wait-for-it.sh:
# curl -O https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh
# chmod +x wait-for-it.sh

ADD ./authservice-linux-amd64 /authservice
ADD ./authservice.yml /
ADD ./wait-for-it.sh /

EXPOSE 28080

# ENTRYPOINT ["/bin/bash", "-c"]
CMD ["/wait-for-it.sh" , "db:3306" , "--strict" , "--timeout=300" , "--" , \
  "/authservice", "-f", "/authservice.yml"]
