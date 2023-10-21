package multi_cluster_controller

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/practice/multi_resource/pkg/apis/multiclusterresource/v1alpha1"
	"github.com/practice/multi_resource/pkg/caches"
	"github.com/practice/multi_resource/pkg/config"
	"github.com/practice/multi_resource/pkg/util"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// MultiClusterHandler 多集群缓存
type MultiClusterHandler struct {
	// masterCluster 主集群，默认使用列表第一个集群作为主集群
	// 如果配置文件有，会赋值，如果设置多个，后遍历者会把前者覆盖
	MasterCluster string
	// 用于缓存需要的多集群信息，key:集群名，config 中定义 value:由各集群初始化后的对象
	RestConfigMap      map[string]*rest.Config
	DynamicClientMap   map[string]dynamic.Interface
	RestMapperMap      map[string]*meta.RESTMapper
	InformerFactoryMap map[string]dynamicinformer.DynamicSharedInformerFactory
	HandlerMap         map[string]*caches.ResourceHandler
	// operator 控制器 client
	client.Client
	// 事件发送器
	EventRecorder record.EventRecorder

	logr.Logger
}

// NewMultiClusterHandlerFromConfig 输入配置文件目录，返回MultiClusterInformer对象
func NewMultiClusterHandlerFromConfig(path string, db *gorm.DB) (*MultiClusterHandler, error) {
	// 解析 config
	sysConfig, err := config.BuildConfig(path)
	if err != nil {
		klog.Error("load config error: ", err)
		return nil, err
	}
	return newMultiClusterHandler(sysConfig.Clusters, db)
}

// newMultiClusterHandler 初始化各集群需要的资源
func newMultiClusterHandler(clusters []config.Cluster, db *gorm.DB) (*MultiClusterHandler, error) {

	if len(clusters) == 0 {
		panic("empty cluster...")
	}

	core := &MultiClusterHandler{
		RestConfigMap:      map[string]*rest.Config{},
		RestMapperMap:      map[string]*meta.RESTMapper{},
		DynamicClientMap:   map[string]dynamic.Interface{},
		InformerFactoryMap: map[string]dynamicinformer.DynamicSharedInformerFactory{},
		HandlerMap:         map[string]*caches.ResourceHandler{},
	}

	// 遍历
	for _, v := range clusters {
		// 如果有主集群
		if v.MetaData.IsMaster {
			core.MasterCluster = v.MetaData.ClusterName
		}
		// 处理需要的初始化
		if v.MetaData.ConfigPath != "" {
			k8sConfig := config.NewK8sConfig(v.MetaData.ConfigPath, v.MetaData.Insecure)
			watcher, dyclient, restConfig := k8sConfig.InitWatchFactoryAndRestConfig()
			// 用于 GVR GVK 转换
			restMapper := k8sConfig.NewRestMapperOrDie()
			// 初始化回调处理函数
			handler := caches.NewResourceHandler(db, *restMapper, v.MetaData.ClusterName)

			// 获取资源哪些资源对象
			for _, vv := range v.MetaData.Resources {
				gvr := util.ParseIntoGvr(vv.RType, "/")
				_, err := watcher.ForResource(gvr).Informer().AddEventHandler(handler)
				if err != nil {
					continue
				}
			}

			// 存入
			core.InformerFactoryMap[v.MetaData.ClusterName] = watcher
			core.HandlerMap[v.MetaData.ClusterName] = handler
			core.DynamicClientMap[v.MetaData.ClusterName] = dyclient
			core.RestConfigMap[v.MetaData.ClusterName] = restConfig
			core.RestMapperMap[v.MetaData.ClusterName] = restMapper
		}
	}

	// 适配，如果初始化后仍没有主集群，默认使用第一个当主集群
	if core.MasterCluster == "" && len(clusters) > 0 {
		core.MasterCluster = clusters[0].MetaData.ClusterName
	}

	return core, nil
}

// StartWorkQueueHandler 启动所有 informerFactory 与 work queue
func (mc *MultiClusterHandler) StartWorkQueueHandler(ctx context.Context) {
	for r := range mc.HandlerMap {
		klog.Infof("[%s] informer watcher start..\n", r)
		mc.HandlerMap[r].Start(ctx)
		mc.InformerFactoryMap[r].Start(wait.NeverStop)
		mc.InformerFactoryMap[r].WaitForCacheSync(wait.NeverStop)
	}
}

// StartOperatorManager 初始化控制器管理器
func (mc *MultiClusterHandler) StartOperatorManager() error {
	logf.SetLogger(zap.New())
	// 1. 初始化管理器，使用
	mgr, err := manager.New(mc.RestConfigMap[mc.MasterCluster],
		manager.Options{
			Logger: logf.Log.WithName("multi-cluster-operator"),
		})

	if err != nil {
		mc.Logger.Error(err, "unable to set up manager")
		return err
	}
	// 2. 安装 CRD 资源对象
	mc.applyCrdToMasterClusterOrDie()

	// 3. 注册进入序列化表
	err = v1alpha1.SchemeBuilder.AddToScheme(mgr.GetScheme())
	if err != nil {
		mc.Logger.Error(err, "unable add schema")
		return err
	}

	// 4. 赋值 operator 需要的 client EventRecorder
	if mc.Client == nil {
		mc.Client = mgr.GetClient()
	}
	if mc.EventRecorder == nil {
		mc.EventRecorder = mgr.GetEventRecorderFor("multi-cluster-operator")
	}
	mc.Logger = mgr.GetLogger()

	if err = builder.ControllerManagedBy(mgr).
		For(&v1alpha1.MultiClusterResource{}).
		Complete(mc); err != nil {
		mc.Logger.Error(err, "unable to create manager")
		return err
	}

	// 5. 启动controller管理器
	if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
		mc.Logger.Error(err, "unable to start manager")
		return err
	}

	return nil
}
