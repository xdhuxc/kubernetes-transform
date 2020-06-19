package service

import (
	"github.com/jinzhu/gorm"

	"github.com/xdhuxc/kubernetes-transform/src/model"
)

type healthService struct {
	db *gorm.DB
}

func newHealthService(db *gorm.DB) *healthService {
	return &healthService{db}
}

func (hcs *healthService) Get() (model.Health, error) {
	return model.Health{
		Message: "kubernetes R-S-T service is OK",
	}, nil
}
