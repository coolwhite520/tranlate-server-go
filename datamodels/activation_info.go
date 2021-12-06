package datamodels

import "time"

type ActivationInfo struct {
	UserName string
	SupportLangList []string
	CreatedDate time.Time
	ExpiredDate time.Time
	MachineId string
}
