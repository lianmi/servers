apps = 'dispatcher' 'chatservice' 'orderservice' 'walletservice'
 
#BINARY='lianmi-admin-server'

LOG_DIR=/root/developments/lianmi/work/logs

.PHONY: run

run: 
	@echo "Mac: make mac, Linux: make linux"
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

.PHONY: admin_linux
admin_linux:
	#GOOS=linux GOARCH="amd64" go build -o dist/gin-vue-admin-linux-amd64 ./cmd/gin-vue-admin
	$(MAKE) -C internal/app/gin-vue-admin/ all
	
.PHONY: admin_mac
admin_mac:
	GOOS=darwin GOARCH="amd64" go build -o dist/gin-vue-admin-darwin-amd64 ./cmd/gin-vue-admin

.PHONY: linux
linux:
	for app in $(apps) ;\
	do \
		GOOS=linux GOARCH="amd64" go build -o dist/$$app-linux-amd64 ./cmd/$$app/; \
	done
	# Make lianmi-admin-server (incl. binary and docker image)
	$(MAKE) -C internal/app/gin-vue-admin/ all
	
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

.PHONY: dart
dart:

	rm -f dart_out/api/proto/auth/*.dart
	protoc --dart_out=dart_out api/proto/auth/AddSlaveDevice.proto \
								api/proto/auth/AuthorizeCode.proto \
								api/proto/auth/GetAllDevices.proto \
								api/proto/auth/Kick.proto \
								api/proto/auth/KickedEvent.proto \
								api/proto/auth/MultiLoginEvent.proto \
								api/proto/auth/Service.proto \
								api/proto/auth/SignIn.proto \
								api/proto/auth/SignOut.proto \
								api/proto/auth/SlaveDeviceAuthEvent.proto \
								api/proto/auth/ValidCode.proto

	rm -f dart_out/api/proto/friends/*.dart
	protoc --dart_out=dart_out api/proto/friends/CancelWatchRequest.proto \
								api/proto/friends/DeleteFriend.proto \
								api/proto/friends/FriendChange.proto \
								api/proto/friends/FriendRequest.proto \
								api/proto/friends/GetFriends.proto \
								api/proto/friends/PsSource.proto \
								api/proto/friends/SyncFriendUsers.proto \
								api/proto/friends/SyncFriends.proto \
								api/proto/friends/SyncUpdateFriendEvent.proto \
								api/proto/friends/UpdateFriend.proto \
								api/proto/friends/WatchRequest.proto

	rm -f dart_out/api/proto/global/*.dart
	protoc --dart_out=dart_out api/proto/global/Global.proto
	

	rm -f dart_out/api/proto/msg/*.dart
	protoc --dart_out=dart_out api/proto/msg/GetOssToken.proto \
	 						   api/proto/msg/MessageAttachType.proto \
							   api/proto/msg/MessagePackage.proto \
							   api/proto/msg/MsgAck.proto \
							   api/proto/msg/MsgTypeEnum.proto \
							   api/proto/msg/RecvCancelMsgEvent.proto \
							   api/proto/msg/RecvMsgEvent.proto \
							   api/proto/msg/SendCancelMsg.proto \
							   api/proto/msg/SendMsg.proto \
							   api/proto/msg/SyncOfflineSysMsgsEvent.proto \
							   api/proto/msg/SyncSendCancelMsgEvent.proto \
							   api/proto/msg/SyncSendMsgEvent.proto \
							   api/proto/msg/grpc.proto

	rm -f dart_out/api/proto/order/*.dart
	protoc --dart_out=dart_out api/proto/order/AddProduct.proto \
								api/proto/order/AddProductEvent.proto \
								api/proto/order/ChangeOrderState.proto \
								api/proto/order/GetPreKeyOrderID.proto \
								api/proto/order/GetPreKeysCount.proto \
								api/proto/order/OPKLimitAlert.proto \
								api/proto/order/OrderPayDoneEvent.proto \
								api/proto/order/PayOrder.proto \
								api/proto/order/Product.proto \
								api/proto/order/QueryProducts.proto \
								api/proto/order/RegisterPreKeys.proto \
								api/proto/order/Service.proto \
								api/proto/order/SoldouProduct.proto \
								api/proto/order/SoldoutProductEvent.proto \
								api/proto/order/SyncGeneralProductsEvent.proto \
								api/proto/order/SyncProductsEvent.proto \
								api/proto/order/SyncWatchEvent.proto \
								api/proto/order/UpdateProduct.proto \
								api/proto/order/UpdateProductEvent.proto \
								api/proto/order/grpcOrder.proto

	rm -f dart_out/api/proto/syn/*.dart
	protoc --dart_out=dart_out api/proto/syn/Sync.proto \
	 							api/proto/syn/SyncDone.proto

	rm -f dart_out/api/proto/team/*.dart
	protoc --dart_out=dart_out api/proto/team/AcceptTeamInvite.proto \
	 api/proto/team/AddTeamManagers.proto \
	 api/proto/team/ApplyTeam.proto \
	 api/proto/team/CheckTeamInvite.proto \
	 api/proto/team/CreateTeam.proto \
	 api/proto/team/DismissTeam.proto \
	 api/proto/team/GetMyTeams.proto \
	 api/proto/team/GetTeam.proto \
	 api/proto/team/GetTeamMembers.proto \
	 api/proto/team/GetTeamMembersPage.proto \
	 api/proto/team/InviteTeamMembers.proto \
	 api/proto/team/LeaveTeam.proto \
	 api/proto/team/MuteTeam.proto \
	 api/proto/team/MuteTeamMember.proto \
	 api/proto/team/PassTeamApply.proto \
	 api/proto/team/PullTeamMembers.proto \
	 api/proto/team/RejectTeamApply.proto \
	 api/proto/team/RejectTeamInvite.proto \
	 api/proto/team/RemoveTeamManagers.proto \
	 api/proto/team/RemoveTeamMembers.proto \
	 api/proto/team/SetNotifyType.proto \
	 api/proto/team/SyncCreateTeam.proto  \
	 api/proto/team/SyncMyTeamsEvent.proto \
	 api/proto/team/TeamInfo.proto \
	 api/proto/team/TransferTeam.proto  \
	 api/proto/team/UpdateMemberInfo.proto \
	 api/proto/team/UpdateMyInfo.proto \
	 api/proto/team/UpdateTeam.proto

	rm -f dart_out/api/proto/user/*.dart
	protoc --dart_out=dart_out api/proto/user/MarkTag.proto \
								api/proto/user/SyncMarkTagEvent.proto \
								api/proto/user/SyncTagsEvent.proto \
								api/proto/user/SyncUpdateProfileEvent.proto \
								api/proto/user/SyncUserProfileEvent.proto \
								api/proto/user/UpdateProfile.proto \
								api/proto/user/UpdateUserProfile.proto \
								api/proto/user/User.proto \
								api/proto/user/Store.proto 

	rm -f dart_out/api/proto/wallet/*.*.dart
	protoc --dart_out=dart_out api/proto/wallet/Alipay.proto \
								api/proto/wallet/Balance.proto \
								api/proto/wallet/ConfirmTransfer.proto \
								api/proto/wallet/Deposit.proto \
								api/proto/wallet/EthReceivedEvent.proto \
								api/proto/wallet/LNMCReceivedEvent.proto  \
								api/proto/wallet/PreTransfer.proto \
								api/proto/wallet/PreWithDraw.proto \
								api/proto/wallet/QueryCommission.proto \
								api/proto/wallet/RegisterWallet.proto \
								api/proto/wallet/SyncCollectionHistory.proto \
								api/proto/wallet/SyncDepositHistory.proto \
								api/proto/wallet/SyncTransferHistory.proto \
								api/proto/wallet/SyncWithdrawHistory.proto \
								api/proto/wallet/TxHashInfo.proto \
								api/proto/wallet/UserSignIn.proto \
								api/proto/wallet/WXPay.proto \
								api/proto/wallet/Wallet.proto \
								api/proto/wallet/WithDraw.proto \
								api/proto/wallet/WithDrawBankCompleteEvent.proto \
								api/proto/wallet/grpc.proto


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
	
	rm -rf dart_out/api

	
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
	# export env_file=/root/developments/lianmi/lm-cloud/servers/deployments/.env
	# echo "LOG_DIR=/root/developments/lianmi/work/logs" > $env_file
	mkdir -p $(LOG_DIR)
	echo $(apps) | tr ' ' '\n' | xargs -i touch "$(LOG_DIR)/{}.log"
	# DIRTY
	touch $(LOG_DIR)/lianmi-admin-server.log
	
	docker-compose -f deployments/docker-compose.yml down
	docker-compose -f deployments/docker-compose.yml up --build -d
all: lint cover docker


.PHONY: docker-clean
docker-clean:
	-docker rmi $(shell docker images --filter "dangling=true" -q --no-trunc)