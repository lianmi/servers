# lianmi-admin-web Frontend Builder Script
# Path: /root/developments/lianmi/lm-cloud/servers/cmd/web/

VERSION?=0.1

STATIC_DEPLOY_PATH?=/root/developments/lianmi/www/admin.lianmi.cloud
DOCKER_IMAGE?=lianmi/admin-web:dist-${VERSION}


.PHONY: dist deploy-only deploy clean

dist: clean
	docker build -f Dockerfile.web -t ${DOCKER_IMAGE} .
	
deploy-only:
	$(eval CID=$(shell docker create ${DOCKER_IMAGE}))

	mkdir -p ${STATIC_DEPLOY_PATH}
	rm -rf ${STATIC_DEPLOY_PATH}/*
	
	docker cp $(CID):/dist ${STATIC_DEPLOY_PATH}
	docker rm $(CID)
    
	mv ${STATIC_DEPLOY_PATH}/dist/* ${STATIC_DEPLOY_PATH}
	rm -rf ${STATIC_DEPLOY_PATH}/dist


deploy: dist deploy-only

clean:
	-docker rmi ${DOCKER_IMAGE}