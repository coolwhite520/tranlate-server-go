package datamodels

//IpTableRecord 黑白名单记录
type IpTableRecord struct {
	Id       int64  `json:"id"`
	Ip       string `json:"ip"`
	Type     string `json:"type"`
	CreateAt string `json:"create_at"`
}
