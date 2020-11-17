package common

const (
	// SERVER_URL = "http://127.0.0.1:28080"
	SERVER_URL = "https://127.0.0.1:28080"

	//http证书路径，按自己实际修改, 注意：最后没有/
	CaCertPath = "/Users/mac/developments/lianmi/lm-cloud/servers/lmSdkClient/ca"

	ENDPOINT_SMSCODE = "/smscode/%s"
	ENDPOINT_LOGIN   = "/login"

	CaPath     = "/Users/mac/developments/lianmi/lm-cloud/servers/lmSdkClient/193"
	BrokerAddr = "192.168.1.193:1883"
	// BrokerAddr = "mqtt.lianmi.cloud:1883"

	RedisAddr = "127.0.0.1:6379"

	// SeedPassword = "socialhahasky"
	SeedPassword = "" //TODO 暂时不动，等准备上线后再统一改

)
