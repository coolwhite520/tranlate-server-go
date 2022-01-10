package services

import "translate-server/datamodels"

type BlackWhiteService interface {
	AddRecord(record datamodels.BlackWhiteRecord) error
	DelRecord(Id int64) error
	QueryRecordsByType(tType int) ([]datamodels.BlackWhiteRecord, error)
}

func NewBlackWhiteService() BlackWhiteService {
	return &blackWhiteService{}
}

type blackWhiteService struct {

}

func (b *blackWhiteService) AddRecord(record datamodels.BlackWhiteRecord) error {
	return nil
}

func (b *blackWhiteService) DelRecord(Id int64) error {
	return nil
}

func (b *blackWhiteService) QueryRecordsByType(tType int) ([]datamodels.BlackWhiteRecord, error) {
	return nil, nil
}