package datamodels

type TransStatus int64

const (
	TransNoRun TransStatus = iota
	TransBeginExtract
	TransExtractFailed
	TransExtractSuccess
	TransBeginTranslate
	TransTranslateFailed
	TransTranslateSuccess
)

func (t TransStatus) String() string {
	switch t {
	case TransNoRun:
		return "上传成功并未开启翻译"
	case TransBeginExtract:
		return "正在抽取文件内容"
	case TransExtractSuccess:
		return "抽取文件内容成功"
	case TransExtractFailed:
		return "抽取文件内容失败"
	case TransBeginTranslate:
		return "正在进行翻译"
	case TransTranslateFailed:
		return "翻译失败"
	case TransTranslateSuccess:
		return "翻译成功"
	default:
		return ""
	}
}

type Record struct {
	Id            int64       `json:"id"`
	Md5           string      `json:"md5"` // 文本的Md5或文件的md5
	Content       string      `json:"content"`
	ContentType   string      `json:"content_type"` // text , application/zip, image/png ,
	TransType     int         `json:"trans_type"`   // 0: 文本 , 1：图片  2: 文档
	OutputContent string      `json:"output_content"`
	SrcLang       string      `json:"src_lang"`
	DesLang       string      `json:"des_lang"`
	FileName      string      `json:"file_name"`
	DirRandId     string      `json:"dir_rand_id"`
	State         TransStatus `json:"state"`
	StateDescribe string      `json:"state_describe"`
	Error         string      `json:"error"`
	UserId        int64       `json:"user_id"`
	CreateAt      string      `json:"create_at"`
}
