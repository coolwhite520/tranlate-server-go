package datamodels

type ActivationInfo struct {
	UserName        string   `json:"user_name"`
	SupportLangList []string `json:"support_lang_list"`
	CreatedAt       string   `json:"created_at"`
	ExpiredAt       string   `json:"expired_at"`
	MachineId       string   `json:"machine_id"`
	Mark            string   `json:"mark"`
}
