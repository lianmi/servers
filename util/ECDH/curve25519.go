package ECDH

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"io"
	"os"
)

var (
	pub string = "C05546A861B953FC1691AB153B6D1ACAFB6396B61166B3CD20612CE753514113"
	pri string = "7107BAEE6CBED449F1BD68A341958C8F776AC417A339387323D71060FBC95987"
)


func demo() {

	var Aprivate, Apublic [32]byte
	//产生随机数
	if _, err := io.ReadFull(rand.Reader, Aprivate[:]); err != nil {
		os.Exit(0)
	}
	curve25519.ScalarBaseMult(&Apublic, &Aprivate)
	fmt.Println("A私钥: ", hex.EncodeToString(Aprivate[:]))
	fmt.Println("A公钥: ", hex.EncodeToString(Apublic[:])) //作为椭圆起点

	var Bprivate, Bpublic [32]byte
	/*
		//产生随机数
		if _, err := io.ReadFull(rand.Reader, Bprivate[:]); err != nil {
			os.Exit(0)
		}
		curve25519.ScalarBaseMult(&Bpublic, &Bprivate)
		fmt.Println("B私钥", hex.EncodeToString(Bprivate[:]))
		fmt.Println("B公钥", hex.EncodeToString(Bpublic[:])) //作为椭圆起点
	*/

	fmt.Println("B私钥: ", pri)
	fmt.Println("B公钥: ", pub) //作为椭圆起点

	priBytesTemp, _ := hex.DecodeString(pri)
	copy(Bprivate[:], priBytesTemp[:]) //可变大小的数组[]byte 转换为固定大小的数组字节
	pubBytesTemp, _ := hex.DecodeString(pub)
	copy(Bpublic[:], pubBytesTemp[:])

	var DH1, DH2 [32]byte

	curve25519.ScalarMult(&DH1, &Aprivate, &Bpublic) //A的私钥加上Ｂ的公钥计算A的key
	fmt.Println("DH1: ", hex.EncodeToString(DH1[:]))

	curve25519.ScalarMult(&DH2, &Bprivate, &Apublic) //B的私钥加上A的公钥计算B的key
	fmt.Println("DH2: ", hex.EncodeToString(DH2[:]))

	if DH1 != DH2 {
		fmt.Println(" DH1 != DH2")
		os.Exit(0)
	}

	// SetAesKey(string(DH1[:]))

	// ciphertext, err := AesCFBEncrypt([]byte("hahaha"), "ZeroPadding")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(0)

	// }
	ciphertext := AesCTR_Encrypt([]byte("hahaha"), DH1[:])

	fmt.Println("密文ciphertext: ", hex.EncodeToString(ciphertext))

	// SetAesKey(string(DH2[:]))
	// plaintext, err := AesCFBDecrypt(ciphertext, "ZeroUnPadding")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(0)

	// }

	plaintext := AesCTR_Decrypt(ciphertext, DH2[:])
	fmt.Println("明文plaintext: ", string(plaintext))
}
