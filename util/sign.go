package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// 生成签名
func GenerateSignature(data string, secretKey string) string {
	hmacSha256 := hmac.New(sha256.New, []byte(secretKey))
	hmacSha256.Write([]byte(data))
	return hex.EncodeToString(hmacSha256.Sum(nil))
}
