package mysql

import (
	"github.com/myoperator/multiclusteroperator/pkg/store"
	"github.com/pkg/errors"
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
	st := &mysqlStoreFactory{
		db: db,
	}
	return st, nil
}
