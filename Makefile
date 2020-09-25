# apps = 'authservice' 'dispatcher' 'chatservice' 'orderservice' 'walletservice'
apps = 'authservice' 'dispatcher' 
# 
.PHONY: run
run: proto wire
	for app in $(apps) ;\
	do \
		 go run ./cmd/$$app -f configs/$$app.yml  & \
	done
.PHONY: wire
wire:
	wire ./...
.PHONY: test
test: mock
	for app in $(apps) ;\
	do \
		go test -v ./internal/app/$$app/... -f `pwd`/configs/$$app.yml -covermode=count -coverprofile=dist/cover-$$app.out ;\
	done
.PHONY: mac
mac:
	for app in $(apps) ;\
	do \
		GOOS=darwin GOARCH="amd64" go build -o dist/$$app-darwin-amd64 ./cmd/$$app/; \
	done
.PHONY: linux
linux:
	for app in $(apps) ;\
	do \
		GOOS=linux GOARCH="amd64" go build -o dist/$$app-linux-amd64 ./cmd/$$app/; \
	done
.PHONY: cover
cover: test
	for app in $(apps) ;\
	do \
		go tool cover -html=dist/cover-$$app.out; \
	done
.PHONY: mock
mock:
	mockery --all
.PHONY: lint
lint:
	golint ./...
.PHONY: proto
proto:

	rm -f api/proto/auth/*.go
	protoc --go_out=plugins=grpc,paths=source_relative:. ./api/proto/auth/*.proto

	rm -f api/proto/friends/*.go
	protoc --go_out=plugins=grpc,paths=source_relative:. ./api/proto/friends/*.proto

	rm -f api/proto/global/*.go
	protoc --go_out=plugins=grpc,paths=source_relative:. ./api/proto/global/*.proto

	rm -f api/proto/msg/*.go
	protoc --go_out=plugins=grpc,paths=source_relative:. ./api/proto/msg/*.proto

	rm -f api/proto/order/*.go
	protoc --go_out=plugins=grpc,paths=source_relative:. ./api/proto/order/*.proto

	rm -f api/proto/syn/*.go
	protoc --go_out=plugins=grpc,paths=source_relative:. ./api/proto/syn/*.proto

	rm -f api/proto/team/*.go
	protoc --go_out=plugins=grpc,paths=source_relative:. ./api/proto/team/*.proto

	rm -f api/proto/user/*.go
	protoc --go_out=plugins=grpc,paths=source_relative:. ./api/proto/user/*.proto

	rm -f api/proto/wallet/*.go
	protoc --go_out=plugins=grpc,paths=source_relative:. ./api/proto/wallet/*.proto

.PHONY: clear
clear:

	rm -f api/proto/auth/*.go
	rm -f api/proto/friends/*.go
	rm -f api/proto/global/*.go
	rm -f api/proto/msg/*.go
	rm -f api/proto/order/*.go
	rm -f api/proto/syn/*.go
	rm -f api/proto/team/*.go
	rm -f api/proto/user/*.go
	rm -f api/proto/wallet/*.go
	
.PHONY: stop
stop:
	docker-compose -f deployments/docker-compose.yml down
.PHONY: dash
dash: # create grafana dashboard
	 for app in $(apps) ;\
	 do \
	 	jsonnet -J ./grafana/grafonnet-lib   -o ./configs/grafana/dashboards/$$app.json  --ext-str app=$$app ./scripts/grafana/dashboard.jsonnet ;\
	 done
.PHONY: pubdash
pubdash:
	 for app in $(apps) ;\
	 do \
	 	jsonnet -J ./grafana/grafonnet-lib  -o ./configs/grafana/dashboards-api/$$app-api.json  --ext-str app=$$app  ./scripts/grafana/dashboard-api.jsonnet ; \
	 	curl -X DELETE --user admin:admin  -H "Content-Type: application/json" 'http://localhost:3000/api/dashboards/db/$$app'; \
	 	curl -x POST --user admin:admin  -H "Content-Type: application/json" --data-binary "@./configs/grafana/dashboards-api/$$app-api.json" http://localhost:3000/api/dashboards/db ; \
	 done
.PHONY: rules
rules:
	for app in $(apps) ;\
	do \
	 	jsonnet  -o ./configs/prometheus/rules/$$app.yml --ext-str app=$$app  ./scripts/prometheus/rules.jsonnet ; \
	done
.PHONY: docker
docker-compose: build dash rules
	for app in $(apps) ;\
	do \
		cp -f dist/$$app-linux-amd64 ./deployments/$$app/; \
	done
	export env_file=/root/developments/lianmi/lm-cloud/servers/deployments/.env
	echo "LOG_DIR=/root/developments/lianmi/work/logs" > $env_file
	docker-compose -f deployments/docker-compose.yml up --build -d
all: lint cover docker
