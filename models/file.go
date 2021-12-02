package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

type File struct {
	Id string `json:"id"`
	SrcLang string `json:"src_lang"`
	DesLang string `json:"des_lang"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Md5 string `json:"md5"`
	OutPutFilePath string `json:"output_file_path"`
	Datetime time.Time `json:"datetime"`
	UserId string `json:"user_id"`
}



func GetFileMd5(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("os Open error")
		return "", err
	}
	md5 := md5.New()
	_, err = io.Copy(md5, file)
	if err != nil {
		fmt.Println("io copy error")
		return "", err
	}
	md5Str := hex.EncodeToString(md5.Sum(nil))
	return md5Str, nil
}

