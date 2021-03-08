# lianmi-admin-server Backend Builder Script
# Path: /root/developments/lianmi/lm-cloud/servers/internal/app/gin-vue-admin

VERSION?=0.1

DOCKER_IMAGE?=lianmi/admin-server:${VERSION}

BINARY="lianmi-admin-server"
GVA="gva"

.PHONY: linux gva docker all check clean

linux: check
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}-linux-amd64
	
gva:
	go env -w GO111MODULE=on
	go env -w GOPROXY=https://goproxy.io,direct
	go build -o ${GVA} cmd/main.go
	
initdb:
	@if [ -f ${GVA} ] ; then ./gva initdb && rm ${GVA} ; fi

docker: linux
	-docker rmi ${DOCKER_IMAGE}
	docker build -t ${DOCKER_IMAGE} .

all: docker

check:
	go fmt ./
	go vet ./
	
clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	@if [ -f ${GVA} ] ; then rm ${GVA} ; fi
	