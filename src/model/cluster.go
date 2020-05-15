package model

import (
	"encoding/json"
)

type Cluster struct {
	BaseModel

	Address string `json:"address" gorm:"address"`
	Token   string `json:"token" gorm:"token"`
	Cloud   string `json:"cloud" gorm:"cloud"`
	Region  string `json:"region" gorm:"region"`
}

func (c *Cluster) TableName() string {
	return "k8s_cluster"
}

func (c *Cluster) String() string {
	if dataInBytes, err := json.Marshal(c); err == nil {
		return string(dataInBytes)
	}

	return ""
}
