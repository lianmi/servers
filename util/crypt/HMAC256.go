package crypt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

//返回Base64编码后的HMAC256
func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
