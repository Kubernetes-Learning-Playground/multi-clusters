package mysql

import (
	"github.com/pkg/errors"
	"github.com/practice/multi_resource/pkg/store"
	"gorm.io/gorm"
)

type mysqlStoreFactory struct {
	db *gorm.DB
}

var _ store.Factory = (*mysqlStoreFactory)(nil)

func (ds *mysqlStoreFactory) Close() error {
	db, err := ds.db.DB()
	if err != nil {
		return errors.Wrap(err, "get gorm db instance failed")
	}
	return db.Close()
}

func (ds *mysqlStoreFactory) GetDB() *gorm.DB {
	return ds.db
}

// NewStoreFactory 创建实例
func NewStoreFactory(db *gorm.DB) (store.Factory, error) {
	store := &mysqlStoreFactory{
		db: db,
	}
	return store, nil
}
