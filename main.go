package main

import (
	"context"
	"flag"
	"github.com/practice/multi_resource/pkg/config"
	"github.com/practice/multi_resource/pkg/httpserver"
	"github.com/practice/multi_resource/pkg/multi_cluster"
	"k8s.io/klog/v2"
)

var (
	dbUser     string // db user
	dbPassword string // db password
	dbEndpoint string // db ip:端口
	dbTable    string // db 表
	configPath string // 配置文件路径
	debugMode  bool   // 是否debug模式
	port       int    // 端口
	healthPort int    // 健康检查端口
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

	// 配置项
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

	// 依赖项
	dp := &config.Dependencies{
		DB: db,
	}

	mch, err := multi_cluster.NewMultiClusterHandlerFromConfig(opt.ConfigPath, db)
	if err != nil {
		klog.Fatal(err)
	}

	// 启动多集群 handler
	mch.Start(ctx)

	errC := make(chan error)

	// 启动httpServer
	go func() {
		klog.Info("httpserver start!! ")
		if err = httpserver.HttpServer(ctx, opt, dp); err != nil {
			errC <- err
		}
	}()

	select {
	case <-ctx.Done():
		klog.Errorf("internal error: %s", ctx.Err())
		break
	case ee := <-errC:
		klog.Errorf("http server internal error: %s", ee)
		break
	}
}
