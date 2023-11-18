package config

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/practice/multi_resource/pkg/util"
	mysql1 "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
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

func (dbc *DbConfig) InitDBOrDie() *gorm.DB {

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbc.User, dbc.Password, dbc.Endpoint, dbc.Database)
	gormdb, err := gorm.Open(mysql1.Open(dsn), &gorm.Config{})
	if err != nil {
		klog.Fatalln(err)
	}
	db, err := gormdb.DB()
	if err != nil {
		klog.Fatalln(err)
	}
	db.SetConnMaxLifetime(time.Minute * 10)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)

	db1, _ := sql.Open("mysql", dsn)
	// 关闭数据库连接
	defer db1.Close()
	driver, err := mysql.WithInstance(db1, &mysql.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s/migrations", util.GetWd()),
		"mysql",
		driver,
	)
	if err != nil {
		klog.Fatal(err)
	}

	// 执行数据库迁移
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		klog.Fatalf("Failed to apply migrations: %v", err)
	}

	// 获取当前数据库迁移版本
	version, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		klog.Fatalf("Failed to get migration version: %v", err)
	}

	klog.Infof("Applied migrations up to version: %v\n", version)

	return gormdb
}
