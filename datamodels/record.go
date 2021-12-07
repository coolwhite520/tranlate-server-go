package datamodels

type File struct {
	Id             string `json:"id"`
	Md5            string `json:"md5"`    // 文本的Md5或文件的md5
	SrcLang        string `json:"src_lang"`
	DesLang        string `json:"des_lang"`
	FileName       string `json:"file_name"`
	FilePath       string `json:"file_path"`

	OutputFilePath string `json:"output_file_path"`
	CreateAt       string `json:"create_at"`
	UserId         string `json:"user_id"`
}
