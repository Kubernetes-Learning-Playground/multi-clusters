package options

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/practice/multi_resource/pkg/store"
	mysqlstore "github.com/practice/multi_resource/pkg/store/mysql"
	"github.com/practice/multi_resource/pkg/util"
	"github.com/spf13/pflag"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"log"
	"time"
)

type MySQLOptions struct {
	Host                  string
	Username              string
	Password              string
	Database              string
	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifeTime time.Duration
}

func NewMySQLOptions() *MySQLOptions {
	return &MySQLOptions{}
}

func (o *MySQLOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Host, "db-endpoint", "127.0.0.1",
		"MySQL service host address. Default to 127.0.0.1.")
	fs.StringVar(&o.Username, "db-user", "root",
		"Username for access to mysql service. Default to root.")
	fs.StringVar(&o.Password, "db-password", "1234567",
		"Password for access to mysql. Default to 1234567.")
	fs.StringVar(&o.Database, "db-database", "testdb",
		"Database name for the server to use. Default to empty.")

	fs.IntVar(&o.MaxIdleConnections, "mysql-max-idle-connections", 100,
		"Maximum idle connections allowed to connect to mysql. Default to 100.")
	fs.IntVar(&o.MaxOpenConnections, "mysql-max-open-connections", 100,
		"Maximum open connections allowed to connect to mysql. Default to 100.")
	fs.DurationVar(&o.MaxConnectionLifeTime, "mysql-max-connection-life-time", time.Duration(10)*time.Second,
		"Maximum connection life time allowed to connect to mysql. Default to 10s.")
}

func (o *MySQLOptions) Complete() error {
	return nil
}

func (o *MySQLOptions) Validate() []error {
	var errs []error
	return errs
}

// NewClient 创建 mysql 客户端并进行 migrate
func (o *MySQLOptions) NewClient() (store.Factory, error) {
	dsn := fmt.Sprintf(`%s:%s@tcp(%s)/%s?charset=utf8&parseTime=%t&loc=%s`,
		o.Username,
		o.Password,
		o.Host,
		o.Database,
		true,
		"Local",
	)
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(o.MaxOpenConnections)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(o.MaxConnectionLifeTime)
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(o.MaxIdleConnections)

	// 执行 migrate
	db1, _ := sql.Open("mysql", dsn)
	// 关闭数据库连接
	defer db1.Close()
	driver, err := migratemysql.WithInstance(db1, &migratemysql.Config{})
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

	return mysqlstore.NewStoreFactory(db)
}