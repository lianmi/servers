package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	// "io/ioutil"
	// "os"
)

// 全局变量
// var privateKey, publicKey []byte

// func init() {
// 	var err error
// 	publicKey, err = ioutil.ReadFile("public.pem")
// 	if err != nil {
// 		os.Exit(-1)
// 	}
// 	privateKey, err = ioutil.ReadFile("private.pem")
// 	if err != nil {
// 		os.Exit(-1)
// 	}
// }

/**
 * @brief  获取RSA公钥长度
 * @param[in]       PubKey				    RSA公钥
 * @return   成功返回 RSA公钥长度，失败返回error	错误信息
 */
func GetPubKeyLen(PubKey []byte) (int, error) {
	if PubKey == nil {
		return 0, errors.New("input arguments error")
	}

	block, _ := pem.Decode(PubKey)
	if block == nil {
		return 0, errors.New("public rsaKey error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return 0, err
	}
	pub := pubInterface.(*rsa.PublicKey)

	return pub.N.BitLen(), nil
}

/**
 * @brief  获取RSA私钥长度
 * @param[in]       PriKey				    RSA私钥
 * @return   成功返回 RSA私钥长度，失败返回error	错误信息
 */
func GetPriKeyLen(PriKey []byte) (int, error) {
	if PriKey == nil {
		return 0, errors.New("input arguments error")
	}

	block, _ := pem.Decode(PriKey)
	if block == nil {
		return 0, errors.New("private rsaKey error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return 0, err
	}

	return priv.N.BitLen(), nil
}

// 加密
func RsaEncrypt(publicKey, origData []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// 解密
func RsaDecrypt(privateKey, ciphertext []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

/*
func main() {
	PubKeyLen, _ := GetPubKeyLen(publicKey)
	fmt.Println("pbulic key len is ", PubKeyLen)

	//获取rsa 私钥长度
	PriKeyLen, _ := GetPriKeyLen(privateKey)
	fmt.Println("private key len is ", PriKeyLen)

	//加密

	body := "eyJidXlVc2VyIjoiaWQxIiwiYnVzaW5lc3NVc2VybmFtZSI6ImlkNTgiLCJwcm9kdWN0SUQiOiJjM2M2YjU4MS01MDVmLTQwMDktODg1OS1kNGZkMWFhNGUxMWYiLCJvcmRlcklEIjpudWxsLCJvcmRpZCI6bnVsbCwidGlja2V0VHlwZSI6MSwic3RyYXdzIjpbeyJibHVlQmFsbHMiOls0XSwiZGFudHVvQmFsbHMiOm51bGwsInJlZEJhbGxzIjpbMTcsMjMsMjQsMjgsMjksMzBdfV0sIm11bHRpcGxlIjoxLCJjb3VudCI6MSwiY29zdCI6Mi4wLCJsb3R0ZXJ5UGljT2JqSUQiOm51bGwsImxvdHRlcnlQaWNIYXNoIjpudWxsfQ=="

	data, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		fmt.Println(err)
		return

	}
	cipherData, err := RsaEncrypt(data)
	if err != nil {
		fmt.Println(err)
		return

	}
	cipherBase64 := base64.StdEncoding.EncodeToString(cipherData)
	fmt.Println("cipherBase64: ", cipherBase64)

	fmt.Println()

	encryped, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		fmt.Println(err)
		return

	}

	encrypedBytes, err := RsaDecrypt(encryped)
	if err != nil {
		fmt.Println(err)
		return

	}
	fmt.Println(string(encrypedBytes))

}
*/
