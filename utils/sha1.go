package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func Sha1(data string) string {
	t := sha1.New()
	io.WriteString(t,data)
	return fmt.Sprintf("%x",t.Sum(nil))
}

func SHA1File(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "",err
	}
	h := sha1.New()
	_, err = io.Copy(h,file)
	if err != nil {
		return "",err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}