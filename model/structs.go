package model

type Record struct {
	Id               string `json:"id"`
	Content          string `json:"content"`
	TranslateContent string `json:"translate_content"`
	ContentMd5       string `json:"content_md5"`
}
