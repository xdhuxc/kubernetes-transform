package client

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/xdhuxc/kubernetes-transform/src/config"
)

func NewMySQLClient(dbConfig config.Database) (*gorm.DB, error) {
	uri := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Name)
	db, err := gorm.Open("mysql", uri)
	if err != nil {
		return nil, err
	}

	if err := db.DB().Ping(); err != nil {
		return nil, err
	}
	if config.GetConfig().Debug {
		db.LogMode(true)
	} else {
		db.LogMode(false)
	}

	db.DB().SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.DB().SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.DB().SetConnMaxLifetime(time.Hour)

	return db, nil
}
