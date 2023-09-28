package main

import (
	"context"
	"flag"
	"github.com/practice/multi_resource/pkg/config"
	"github.com/practice/multi_resource/pkg/httpserver"
	"github.com/practice/multi_resource/pkg/multi_cluster_controller"
	"github.com/practice/multi_resource/pkg/util"
	"k8s.io/klog/v2"
)

var (
	dbUser     string // db user
	dbPassword string // db password
	dbEndpoint string // db ip:端口
	dbDatabase string // db
	configPath string // 配置文件路径
	debugMode  bool   // 是否debug模式
	port       int    // 端口
	ctlPort    int    // ctl 需要的端口
	healthPort int    // 健康检查端口
)

func main() {

	flag.StringVar(&dbUser, "db-user", "root", "db user for project")
	flag.StringVar(&dbPassword, "db-password", "1234567", "db password for project")
	flag.StringVar(&dbEndpoint, "db-endpoint", "127.0.0.1:3306", "db endpoint for project")
	flag.StringVar(&dbDatabase, "db-database", "testdb", "db table for project")
	flag.StringVar(&configPath, "config", "./config.yaml", "kubeconfig path for k8s cluster")
	flag.BoolVar(&debugMode, "debug-mode", false, "whether to use debug mode")
	flag.IntVar(&port, "server-port", 8888, "")
	flag.IntVar(&healthPort, "health-check-port", 29999, "")
	flag.IntVar(&ctlPort, "ctl-port", 31888, "")
	flag.Parse()

	// 配置项
	opt := &config.Options{
		User:       dbUser,
		Password:   dbPassword,
		Endpoint:   dbEndpoint,
		Database:   dbDatabase,
		Port:       port,
		ConfigPath: configPath,
		HealthPort: healthPort,
		CtlPort:    ctlPort,
		DebugMode:  debugMode,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建 .multi-cluster-operator config 文件
	config.CreateCtlFile(opt)

	// 初始化数据库
	db := config.NewDbConfig(opt).InitDB()

	// 依赖项
	dp := &config.Dependencies{
		DB: db,
	}

	mch, err := multi_cluster_controller.NewMultiClusterHandlerFromConfig(opt.ConfigPath, db)
	if err != nil {
		klog.Fatal(err)
	}

	// 启动多集群 handler
	mch.StartWorkQueueHandler(ctx)

	errC := make(chan error)

	go func() {
		defer util.HandleCrash()
		klog.Info("httpserver start...")
		if err = httpserver.HttpServer(ctx, opt, dp); err != nil {
			errC <- err
		}
	}()

	go func() {
		defer util.HandleCrash()
		klog.Info("operator manager start...")

		if err = mch.StartOperatorManager(); err != nil {
			errC <- err
		}
	}()

	select {
	case <-ctx.Done():
		klog.Errorf("internal error: %s", ctx.Err())
		break
	case ee := <-errC:
		klog.Errorf("http server or operator manager internal error: %s", ee)
		break
	}
}
