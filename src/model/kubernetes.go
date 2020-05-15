package model

import (
	"encoding/json"
)

type Resource struct {
	BaseModel

	Kind      string `json:"kind" gorm:"kind"`
	Namespace string `json:"namespace" gorm:"namespace"`
	Json      string `json:"json" gorm:"json"`
	Yaml      string `json:"yaml" gorm:"yaml"`

	IsCurrentUpdate bool `json:"-" gorm:"is_current_update"`
}

func (r *Resource) TableName() string {
	return "k8s_resource"
}

func (r *Resource) String() string {
	if dataInBytes, err := json.Marshal(r); err == nil {
		return string(dataInBytes)
	}

	return ""
}

type Exclusion struct {
	Name       string
	Exclusions []string
}

func (e *Exclusion) String() string {
	if dataInBytes, err := json.Marshal(e); err == nil {
		return string(dataInBytes)
	}

	return ""
}
