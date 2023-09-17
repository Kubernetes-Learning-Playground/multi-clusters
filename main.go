package main

import (
	"context"
	"flag"
	"github.com/practice/multi_resource/pkg/caches"
	"github.com/practice/multi_resource/pkg/config"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
)

var (
	dbUser     string // db user
	dbPassword string // db password
	dbEndpoint string // db ip:端口
	dbTable    string // db 表
	// FIXME: 目前暂时使用项目传入方式，之后要适配多集群，需要一个全局配置文件config.yaml
	kubeconfigPath string
	debugMode      bool // 是否debug模式
	port           int  // 端口
	healthPort     int  // 健康检查端口
)

func main() {

	flag.StringVar(&dbUser, "db-user", "root", "db user for project")
	flag.StringVar(&dbPassword, "db-password", "1234567", "db password for project")
	flag.StringVar(&dbEndpoint, "db-endpoint", "127.0.0.1:3306", "db endpoint for project")
	flag.StringVar(&dbTable, "db-table", "testdb", "db table for project")
	flag.StringVar(&kubeconfigPath, "kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "kubeconfig path for k8s cluster")
	flag.BoolVar(&debugMode, "debug-mode", false, "whether to use debug mode")
	flag.IntVar(&port, "server-port", 8888, "")
	flag.IntVar(&healthPort, "health-check-port", 29999, "")
	flag.Parse()

	// 配置文件
	opt := &config.Options{
		User:           dbUser,
		Password:       dbPassword,
		Endpoint:       dbEndpoint,
		Table:          dbTable,
		Port:           port,
		KubeConfigPath: kubeconfigPath,
		HealthPort:     healthPort,
		DebugMode:      debugMode,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化数据库
	db := config.NewDbConfig(opt).InitDB()

	// 动态客户端InformerFactory
	k8sConfig := config.NewK8sConfig(opt)
	watcher := k8sConfig.InitWatchFactory()

	// 用于 GVR GVK 转换
	restMapper := k8sConfig.RestMapper()

	// 初始化回调处理函数
	handler := caches.NewResourceHandler(db, *restMapper)

	// TODO: 目前使用代码加入的方式，未来需要使用配置文件方式自定义需要的资源對象
	// TODO: 可以使用配置文件，解析类似 "apps/v1/deployments" 这种字段
	watcher.ForResource(
		schema.GroupVersionResource{Version: "v1", Resource: "pods"}).Informer().AddEventHandler(handler)
	watcher.ForResource(
		schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}).Informer().AddEventHandler(handler)
	watcher.ForResource(
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}).Informer().AddEventHandler(handler)

	// 启动
	handler.Start(ctx)
	klog.Info("informer watcher start...")
	watcher.Start(wait.NeverStop)
	watcher.WaitForCacheSync(wait.NeverStop)

	select {
	case <-ctx.Done():
		break
	case <-wait.NeverStop:
		break
	}
}
