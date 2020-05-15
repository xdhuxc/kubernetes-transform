package model

import "time"

type BaseModel struct {
	ID   int64  `json:"id" gorm:"id"`
	UUID string `json:"uuid" gorm:"uuid"`
	Name string `json:"name" gorm:"name"`

	Description string    `json:"description" gorm:"description"`
	CreateTime  time.Time `json:"create_time" gorm:"create_time"`
	UpdateTime  time.Time `json:"update_time" gorm:"update_time"`
}
