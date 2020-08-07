package crypt

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	keyPrex = "icoin-sky"
)

// 对字符串进行 sha1 计算
func Sha1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

// 对数据进行md5计算
func MD5(byteMessage []byte) string {
	h := md5.New()
	h.Write(byteMessage)
	return hex.EncodeToString(h.Sum(nil))
}

func HamSha1(data string, key []byte) string {
	hmac := hmac.New(sha1.New, key)
	hmac.Write([]byte(data))

	return base64.StdEncoding.EncodeToString(hmac.Sum(nil))
}

// 生成一个32位的access Key
func GeneratePrivateKey(name string) string {

	_UUID := uuid.NewV3(uuid.NamespaceDNS, keyPrex+name)
	uuID := _UUID.String()
	return strings.Replace(uuID, "-", "", -1)

}

//生成一个32位的ApiSecret
func GenerateSecret(key, mnemonic string) string {
	t := time.Now().Unix() * 1000
	ts := strconv.FormatInt(t, 10)
	_UUID := uuid.NewV3(uuid.NamespaceDNS, keyPrex+key+ts+mnemonic)
	uuID := _UUID.String()

	return strings.Replace(uuID, "-", "", -1)
}

// Sign request with client secret using HMAC-SHA256
// args should be ordered URI format
// API 签名
// 签名前准备的数据如下：
// HTTP_METHOD + HTTP_REQUEST_URI + TIMESTAMP + POST_BODY
// 连接完成后，进行 Base64 编码，对编码后的数据进行 HMAC-SHA256 签名，并对签名进行二次 Base64 编码
func Sign(method, uri, ts, args, secretKey string) string {
	prep := method + uri + ts + args
	//第一次Base64编码
	b64prep := base64.StdEncoding.EncodeToString([]byte(prep))

	// mac := hmac.New(sha1.New, []byte(key))
	// mac.Write(b64prep)

	// hmac_prep := mac.Sum(nil)

	// 签名
	return ComputeHmac256(b64prep, secretKey)
}

func Hmac5Sign(secretKey, body string) []byte {
	m := hmac.New(sha1.New, []byte(secretKey))
	m.Write([]byte(body))
	return m.Sum(nil)
}
