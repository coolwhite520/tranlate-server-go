package structs

type TransStatus int64


const UploadDir = "./data/uploads"
const OutputDir = "./data/outputs"

const (
	TransNoRun TransStatus = iota
	TransBeginTranslate
	TransTranslateFailed
	TransTranslateSuccess
)

func (t TransStatus) String() string {
	switch t {
	case TransNoRun:
		return "上传成功未翻译"
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

// Record 记录用户翻译记录
type Record struct {
	Id            int64       `json:"id"`
	Sha1          string      `json:"sha1"` // 文本或文件的sha1
	Content       string      `json:"content"`
	ContentType   string      `json:"content_type"` // text , application/zip, image/png ,
	TransType     int         `json:"trans_type"`   // 0: 文本 , 1：图片  2: 文档
	OutputContent string      `json:"output_content"`
	SrcLang       string      `json:"src_lang"`
	DesLang       string      `json:"des_lang"`
	FileName      string      `json:"file_name"`   // 去掉后缀名后的结果
	FileExt       string      `json:"file_ext"`
	DirRandId     string      `json:"dir_rand_id"`
	Progress      int         `json:"progress"`
	State         TransStatus `json:"state"`
	StateDescribe string      `json:"state_describe"`
	Error         string      `json:"error"`
	UserId        int64       `json:"user_id"`
	OutFileExt    string      `json:"out_file_ext"` // 输出文件扩展名
	StartAt       string      `json:"start_at"`
	EndAt         string      `json:"end_at"`
	CreateAt      string      `json:"create_at"`   //
}

type RecordEx struct {
	Record
	UserName string `json:"user_name"`
}
