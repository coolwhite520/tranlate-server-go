package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

func GenerateHmacSign(data, key []byte) string {
	 mac := hmac.New(sha1.New, key)
	 mac.Write(data)
	 expectedMAC := mac.Sum(nil)
	 return base64.StdEncoding.EncodeToString(expectedMAC)
}
//signature := generateSign([]byte(data), []byte("123"))
