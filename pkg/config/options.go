package config

import "gorm.io/gorm"

type Options struct {
	User       string
	Password   string
	Endpoint   string
	Table      string
	ConfigPath string
	Port       int
	HealthPort int
	DebugMode  bool
}

type Dependencies struct {
	DB *gorm.DB
}
