package common

const (
	// SERVER_URL = "http://127.0.0.1:28080"
	SERVER_URL = "https://api.lianmi.cloud"

	ENDPOINT_SMSCODE = "/smscode/%s"
	ENDPOINT_LOGIN   = "/login"

	//http证书路径，按自己实际修改, 注意：最后没有/
	CaPath = "/Users/mac/developments/lianmi/lm-cloud/servers/lmSdkClient/ca"

	BrokerAddr = "mqtt.lianmi.cloud:1883"

	RedisAddr = "127.0.0.1:6379"

	// SeedPassword = "socialhahasky"
	SeedPassword = "" //TODO 暂时不动，等准备上线后再统一改

)
