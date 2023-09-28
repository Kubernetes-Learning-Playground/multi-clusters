package config

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

type DbConfig struct {
	User     string
	Password string
	Endpoint string
	Database string
}

func NewDbConfig(opt *Options) *DbConfig {
	return &DbConfig{
		User:     opt.User,
		Password: opt.Password,
		Endpoint: opt.Endpoint,
		Database: opt.Database,
	}
}

func (dbc *DbConfig) InitDB() *gorm.DB {

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbc.User, dbc.Password, dbc.Endpoint, dbc.Database)
	gormdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	db, err := gormdb.DB()
	if err != nil {
		log.Fatalln(err)
	}
	db.SetConnMaxLifetime(time.Minute * 10)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)

	return gormdb
}
