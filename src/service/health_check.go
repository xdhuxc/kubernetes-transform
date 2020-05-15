package service

import (
	"github.com/jinzhu/gorm"

	"github.com/xdhuxc/kubernetes-transform/src/model"
)

type healthCheckService struct {
	db *gorm.DB
}

func newHealthCheckService(db *gorm.DB) *healthCheckService {
	return &healthCheckService{db}
}

func (hcs *healthCheckService) Get() (model.HealthCheck, error) {
	return model.HealthCheck{
		Message: "kubernetes R-S-T service is OK",
	}, nil
}
