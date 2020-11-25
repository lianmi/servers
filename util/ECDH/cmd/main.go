package main

import (
	"encoding/hex"
	"fmt"
	"github.com/lianmi/servers/util/ECDH"
	"golang.org/x/crypto/curve25519"
)

func main() {
	Aprivate, Apublic := ECDH.GetCurve25519KeypPair()
	fmt.Println("A私钥: ", hex.EncodeToString(Aprivate[:]))
	fmt.Println("A公钥: ", hex.EncodeToString(Apublic[:])) //作为椭圆起点

	Bprivate, Bpublic := ECDH.GetCurve25519KeypPair()
	fmt.Println("B私钥: ", hex.EncodeToString(Bprivate[:]))
	fmt.Println("B公钥: ", hex.EncodeToString(Bpublic[:])) //作为椭圆起点

	var DH1, DH2 [32]byte

	curve25519.ScalarMult(&DH1, &Aprivate, &Bpublic) //A的私钥加上Ｂ的公钥计算A的key
	fmt.Println("DH1: ", hex.EncodeToString(DH1[:]))

	curve25519.ScalarMult(&DH2, &Bprivate, &Apublic) //B的私钥加上A的公钥计算B的key
	fmt.Println("DH2: ", hex.EncodeToString(DH2[:]))

	fmt.Println(" DH1 == DH2 ? : ", DH1 == DH2)

	ciphertext := ECDH.AesCTR_Encrypt([]byte("hahaha1234566"), DH1[:])

	fmt.Println("密文ciphertext: ", hex.EncodeToString(ciphertext))

	plaintext := ECDH.AesCTR_Decrypt(ciphertext, DH2[:])
	fmt.Println("明文plaintext: ", string(plaintext))
}
