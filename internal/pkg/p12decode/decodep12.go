package main

import (
	// "crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/crypto/pkcs12"
	"io/ioutil"
	"log"

	// "crypto/rsa"
	"crypto/x509"
	// "encoding/pem"
)

const CertDateBase64 = "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"

func PrintPretty(i interface{}) {
	data, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}

func LoadCert(path, password string) (apiCert *x509.Certificate, err error) {
	if len(path) == 0 {
		return nil, errors.New("Not found cert file")
	}

	certData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	//输出为base64字符串
	certDataBase := base64.StdEncoding.EncodeToString(certData)
	fmt.Printf("certDataBase: %s\n", certDataBase)

	blocks, err := pkcs12.ToPEM(certData, password)

	if err != nil {
		return nil, err
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	//解析出腾讯官方证书
	apiCertBlock, _ := pem.Decode(pemData)
	if apiCertBlock == nil {
		return nil, errors.New("Pem Decode Failed")
	}
	apiCert, err = x509.ParseCertificate(apiCertBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return apiCert, nil
}

func LoadCertFromBase64(base64Str, password string) (apiCert *x509.Certificate, err error) {
	certData, err := base64.StdEncoding.DecodeString(base64Str)

	blocks, err := pkcs12.ToPEM(certData, password)

	if err != nil {
		return nil, err
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	//解析出腾讯官方证书
	apiCertBlock, _ := pem.Decode(pemData)
	if apiCertBlock == nil {
		return nil, errors.New("Pem Decode Failed")
	}
	apiCert, err = x509.ParseCertificate(apiCertBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return apiCert, nil
}

func main() {
	mchId := "1604757586"

	// 加载需要的证书
	// cert, err := LoadCert("/Users/mac/developments/lianmi/lm-cloud/wxpaydemo/apiclient_cert.p12", mchId)
	cert, err := LoadCertFromBase64(CertDateBase64, mchId)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("LoadCert succeed")
		fmt.Println("cert.PublicKey: ", cert.PublicKey)
		// PrintPretty(cert)

	}

}
