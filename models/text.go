package models

import "time"

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

