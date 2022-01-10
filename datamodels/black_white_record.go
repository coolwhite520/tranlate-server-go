package datamodels

//BlackWhiteRecord 黑白名单记录
type BlackWhiteRecord struct {
	Id       int64  `json:"id"`
	Ip       string `json:"ip"`
	Type     int    `json:"type"` //0: white 1: black
	CreateAt string `json:"create_at"`
}
