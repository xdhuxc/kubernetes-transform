package client

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/xdhuxc/kubernetes-transform/src/config"
)

func NewMySQLClient(dbc config.Database) (*gorm.DB, error) {
	uri := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbc.User,
		dbc.Password,
		dbc.Host,
		dbc.Name)
	db, err := gorm.Open("mysql", uri)
	if err != nil {
		return nil, err
	}

	if err := db.DB().Ping(); err != nil {
		return nil, err
	}

	db.LogMode(dbc.Log)
	db.DB().SetMaxIdleConns(dbc.MaxIdleConns)
	db.DB().SetMaxOpenConns(dbc.MaxOpenConns)
	db.DB().SetConnMaxLifetime(time.Hour)

	return db, nil
}
