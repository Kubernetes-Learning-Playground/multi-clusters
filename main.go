package main

import (
	"context"
	"flag"
	"github.com/practice/multi_resource/pkg/config"
	"github.com/practice/multi_resource/pkg/multi_cluster"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var (
	dbUser     string // db user
	dbPassword string // db password
	dbEndpoint string // db ip:端口
	dbTable    string // db 表
	configPath string
	debugMode  bool // 是否debug模式
	port       int  // 端口
	healthPort int  // 健康检查端口
)

func main() {

	flag.StringVar(&dbUser, "db-user", "root", "db user for project")
	flag.StringVar(&dbPassword, "db-password", "1234567", "db password for project")
	flag.StringVar(&dbEndpoint, "db-endpoint", "127.0.0.1:3306", "db endpoint for project")
	flag.StringVar(&dbTable, "db-table", "testdb", "db table for project")
	flag.StringVar(&configPath, "config", "./config.yaml", "kubeconfig path for k8s cluster")
	flag.BoolVar(&debugMode, "debug-mode", false, "whether to use debug mode")
	flag.IntVar(&port, "server-port", 8888, "")
	flag.IntVar(&healthPort, "health-check-port", 29999, "")
	flag.Parse()

	// 配置文件
	opt := &config.Options{
		User:       dbUser,
		Password:   dbPassword,
		Endpoint:   dbEndpoint,
		Table:      dbTable,
		Port:       port,
		ConfigPath: configPath,
		HealthPort: healthPort,
		DebugMode:  debugMode,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化数据库
	db := config.NewDbConfig(opt).InitDB()

	mch, err := multi_cluster.NewMultiClusterHandlerFromConfig(opt.ConfigPath, db)
	if err != nil {
		klog.Fatal(err)
	}

	mch.Start(ctx)

	select {
	case <-ctx.Done():
		break
	case <-wait.NeverStop:
		break
	}
}
