package store

import "gorm.io/gorm"

// Factory 存储接口，目前使用 gorm.DB 实例实现
type Factory interface {
	Close() error
	GetDB() *gorm.DB
}
