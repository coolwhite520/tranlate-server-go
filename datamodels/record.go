package datamodels

type TransState int64

const (
	TransNoRun TransState = iota
	TransBegin
	TransRunning
	TransError
	TransSuccess
)

type Record struct {
	Id             int64      `json:"id"`
	Md5            string     `json:"md5"`        // 文本的Md5或文件的md5
	Content        string     `json:"content"`
	ContentType    string     `json:"content_type"` // text , image/png , application/zip
	OutputContent  string     `json:"output_content"`
	SrcLang        string     `json:"src_lang"`
	DesLang        string     `json:"des_lang"`
	FilePath       string     `json:"file_path"`
	OutputFilePath string     `json:"output_file_path"`
	CreateAt       string     `json:"create_at"`
	State          TransState `json:"state"`
	StateDescribe  string     `json:"state_describe"`
	UserId         int64      `json:"user_id"`
}
