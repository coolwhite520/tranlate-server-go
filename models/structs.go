package models

import "time"

type User struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Password string `json:"password"`
	IsSuper bool `json:"is_super"`
}

type Text struct {
	Id string `json:"id"`
	SrcLang string `json:"src_lang"`
	DesLang string `json:"des_lang"`
	Content string `json:"content"`
	TransContent string `json:"trans_content"`
	Md5 string `json:"md5"`
	Datetime time.Time `json:"datetime"`
	UserId string `json:"user_id"`
}

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