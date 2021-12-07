package datamodels

type Activation struct {
	UserName          string   `json:"user_name"`
	SupportLangList   []string `json:"support_lang_list"`      // 英文简称列表
	SupportLangListCn []string `json:"support_lang_list_cn"`   // 中文简称列表
	CreatedAt         string   `json:"created_at"`
	ExpiredAt         string   `json:"expired_at"`
	MachineId         string   `json:"machine_id"`
	Mark              string   `json:"mark"`
}
