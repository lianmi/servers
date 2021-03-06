/*

https://signin.aliyun.com/1230446857465673.onaliyun.com/login.htm

*/

package sts

import (
	"log"
	"testing"
)

var (
	TempKey    = "LTAI4G3o4sECdSBsD7rGLmCs"                    //固定的
	TempSecret = "0XmB9tLOBLhmjIcM6CrBv2PHfnoDa8"              //固定的
	RoleAcs    = "acs:ram::1230446857465673:role/ipfsuploader" //后台里的RAM角色
)

func TestGenerateSignatureUrl(t *testing.T) {
	client := NewStsClient(TempKey, TempSecret, RoleAcs)

	url, err := client.GenerateSignatureUrl("client", "1800", "")
	if err != nil {
		t.Error(err)
	}

	data, err := client.GetStsResponse(url)
	if err != nil {
		t.Error(err)
	}

	log.Println("result:", string(data))
	/*
		result: {
			"RequestId":"B9A9645F-003C-4203-99A9-E57463D9F2F5",
			"AssumedRoleUser":{
				"Arn":"acs:ram::1230446857465673:role/ipfsuploader/client",
				"AssumedRoleId":"359775758821401491:client"
			},
			"Credentials":{
				"SecurityToken":"CAIS8QF1q6Ft5B2yfSjIr5eHDejxm45ZzYiRNGLcgkw6S7dEn4SYhzz2IH1Fe3ZtBu0Wvv42mGhR6vcblq94T55IQ1CckHn0CUIRo22beIPkl5Gfz95t0e+IewW6Dxr8w7WhAYHQR8/cffGAck3NkjQJr5LxaTSlWS7OU/TL8+kFCO4aRQ6ldzFLKc5LLw950q8gOGDWKOymP2yB4AOSLjIx4FEk1T8hufngnpPBtEWFtjCglL9J/baWC4O/csxhMK14V9qIx+FsfsLDqnUNukcVqfgr3PweoGuf543MWkM14g2IKPfM9tpmIAJjdgmMmRj3JgeWGoABLmgZ3Vg641t3o68K7LfwHMOw7t+h5zfAUzSnohsHTaK4iqIpVmeatqAKbZ59QP/paHFC4WFihtglbyBcrtqz/aIOiKqalI5th9OdHCr/Oj27eGo/cFtOrIA8k6yigQJU45SO6nulRSthnQEl+KjWbSbSgKP4UbETZsT+n63tHTQ=",
				"AccessKeyId":"STS.NT2FREvQxzJz6DmfLoG8hpA3e",
				"AccessKeySecret":"AbVNm43uRPYupGiUSW1kszDjNSbWtgSxZTUFtdW8x7cj",
				"Expiration":"2020-08-12T09:23:21Z"
			}
		}
	*/

}
