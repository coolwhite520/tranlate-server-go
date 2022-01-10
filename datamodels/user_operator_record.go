package datamodels

// UserOperatorRecord 记录用户登录 登出等操作
type UserOperatorRecord struct {
	Id       int64  `json:"id" form:"id"`
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
	Ip       string `json:"ip"`
	Operator string `json:"operator"`
	CreateAt string `json:"create_at"`
}
