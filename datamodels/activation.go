package datamodels

type SupportLang struct {
	EnName string `json:"en_name"`
	CnName string `json:"cn_name"`
}

type Activation struct {
	UserName          string   `json:"user_name"`
	SupportLangList   []SupportLang `json:"support_lang_list"`      // 英文简称列表
	CreatedAt         string   `json:"created_at"`
	ExpiredAt         string   `json:"expired_at"`
	MachineId         string   `json:"machine_id"`
	Mark              string   `json:"mark"`
}
