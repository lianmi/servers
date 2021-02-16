package main

import (
	"encoding/hex"
	"fmt"

	"github.com/lianmi/servers/util/ECDH"
	"golang.org/x/crypto/curve25519"
)

/*
服务端私钥:  b3e7a1b8fa4d7e958eecbc72f8d95e667889b787a83e8be576523aefb82ba507
服务端公钥:  36c02735d5500646e48a10da640713dcc3382347ab7ee2fc15244bbe38270178
*/
func main() {
	Aprivate, Apublic := ECDH.GetCurve25519KeypPair()
	fmt.Println("服务端私钥: ", hex.EncodeToString(Aprivate[:]))
	fmt.Println("服务端公钥: ", hex.EncodeToString(Apublic[:])) //作为椭圆起点

	Bprivate, Bpublic := ECDH.GetCurve25519KeypPair()
	fmt.Println("SDK私钥: ", hex.EncodeToString(Bprivate[:]))
	fmt.Println("SDK公钥: ", hex.EncodeToString(Bpublic[:])) //作为椭圆起点

	var DH_Server, DH_Sdk [32]byte

	curve25519.ScalarMult(&DH_Server, &Aprivate, &Bpublic) //A的私钥加上Ｂ的公钥计算A的key
	fmt.Println("DH_Server: ", hex.EncodeToString(DH_Server[:]))

	curve25519.ScalarMult(&DH_Sdk, &Bprivate, &Apublic) //B的私钥加上A的公钥计算B的key
	fmt.Println("DH_Sdk: ", hex.EncodeToString(DH_Sdk[:]))

	fmt.Println(" DH_Server == DH_Sdk ? : ", DH_Server == DH_Sdk)

	ciphertext := ECDH.AesCTR_Encrypt([]byte("hahaha1234566"), DH_Server[:])

	fmt.Println("密文ciphertext: ", hex.EncodeToString(ciphertext))

	plaintext := ECDH.AesCTR_Decrypt(ciphertext, DH_Sdk[:])
	fmt.Println("明文plaintext: ", string(plaintext))
}
