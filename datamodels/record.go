package datamodels

type TransStatus int64

const (
	TransSuccess TransStatus = iota
	TransNoRun
	TransBegin
	TransRunning
	TransError
)

func (t TransStatus) String() string {
	switch t {
	case TransNoRun:
		return "上传并未启动翻译"
	case TransBegin:
		return "开始对文件进行翻译"
	case TransRunning:
		return "正在对文件进行翻译"
	case TransError:
		return "翻译失败"
	case TransSuccess:
		return "翻译成功"
	default:
		return ""
	}
}

type Record struct {
	Id            int64       `json:"id"`
	Md5           string      `json:"md5"` // 文本的Md5或文件的md5
	Content       string      `json:"content"`
	ContentType   string      `json:"content_type"` // text , image/png , application/zip
	OutputContent string      `json:"output_content"`
	SrcLang       string      `json:"src_lang"`
	DesLang       string      `json:"des_lang"`
	FileName      string      `json:"file_name"`
	FileSrcDir    string      `json:"file_src_dir"` // 文件的原始路径，也就是上传后的路径
	FileDesDir    string      `json:"file_des_dir"` // 文件的目标路径，也就是翻译后的输入文档路径
	State         TransStatus `json:"state"`
	StateDescribe string      `json:"state_describe"`
	Error         string      `json:"error"`
	UserId        int64       `json:"user_id"`
	CreateAt      string      `json:"create_at"`
}
