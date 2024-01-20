package multi_cluster_controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	v1alpha12 "github.com/myoperator/multiclusteroperator/pkg/apis/multicluster/v1alpha1"
	"github.com/myoperator/multiclusteroperator/pkg/apis/multiclusterresource/v1alpha1"
	"github.com/myoperator/multiclusteroperator/pkg/caches"
	"github.com/myoperator/multiclusteroperator/pkg/config"
	"github.com/myoperator/multiclusteroperator/pkg/kubectl_client"
	"github.com/myoperator/multiclusteroperator/pkg/util"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
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

// MultiClusterHandler 多集群控制器实例
type MultiClusterHandler struct {
	// MasterCluster 主集群，默认使用列表第一个集群作为主集群
	// 如果配置文件有，会赋值，如果设置多个，后遍历者会把前者覆盖
	MasterCluster string
	// MasterClusterKubeConfigPath 主集群的 kubeconfig 路径
	MasterClusterKubeConfigPath string
	// 用于缓存需要的多集群信息，key:集群名，config 中定义 value:由各集群初始化后的对象
	RestConfigMap      map[string]*rest.Config
	DynamicClientMap   map[string]dynamic.Interface
	KubectlClientMap   map[string]*kubectl_client.KubectlClient
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
		KubectlClientMap:   map[string]*kubectl_client.KubectlClient{},
	}

	// 遍历
	for _, v := range clusters {
		// 如果有主集群
		if v.MetaData.IsMaster {
			core.MasterCluster = v.MetaData.ClusterName
			core.MasterClusterKubeConfigPath = v.MetaData.ConfigPath
		}
		// 处理需要的初始化
		if v.MetaData.ConfigPath != "" {
			k8sConfig := config.NewK8sConfig(v.MetaData.ConfigPath, v.MetaData.Insecure)
			watcher, dyclient, restConfig := k8sConfig.InitWatchFactoryAndRestConfig()
			// 用于 GVR GVK 转换
			restMapper := k8sConfig.NewRestMapperOrDie()
			// 初始化回调处理函数
			handler := caches.NewResourceHandler(db, *restMapper, v.MetaData.ClusterName)
			kubectlClient := kubectl_client.NewKubectlManagerOrDie(restConfig)

			// TODO: 使用 discovery Client 解析所有 gvr
			// 获取所有资源的 GVR
			apiResources, err := kubectlClient.DiscoveryClient.ServerPreferredResources()
			if err != nil {
				klog.Fatalf("Error getting API resources: %s", err)
			}

			// 输出所有资源的 GVR 加入 handler
			for _, apiResourceList := range apiResources {
				for _, apiResource := range apiResourceList.APIResources {
					groupVersion, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
					if err != nil {
						klog.Errorf("Error parsing GroupVersion: %v", err)
						continue
					}

					if groupVersion.Group == "" {
						groupVersion.Group = "core"
					}

					klog.Infof("GVR: %v/%v/%v\n", groupVersion.Group, groupVersion.Version, apiResource.Name)

					// FIXME: 如果不自定义，这些 group resources 会有异想不到的 bug
					if groupVersion.Group == "metrics.k8s.io" || groupVersion.Group == "authentication.k8s.io" || groupVersion.Group == "authorization.k8s.io" {
						continue
					}
					if groupVersion.Group == "policy" || groupVersion.Group == "apiextensions.k8s.io" || groupVersion.Group == "kueue.x-k8s.io" {
						continue
					}
					if apiResource.Name == "bindings" || apiResource.Name == "componentstatuses" {
						continue
					}

					gvr := util.ParseIntoGvr(fmt.Sprintf("%v/%v/%v", groupVersion.Group, groupVersion.Version, apiResource.Name), "/")
					_, err = watcher.ForResource(gvr).Informer().AddEventHandler(handler)
					if err != nil {
						continue
					}
				}
			}

			// 废弃原本使用配置文件传入的方式
			//for _, vv := range v.MetaData.Resources {
			//	gvr := util.ParseIntoGvr(vv.RType, "/")
			//	_, err := watcher.ForResource(gvr).Informer().AddEventHandler(handler)
			//	if err != nil {
			//		continue
			//	}
			//}

			// 存入
			core.InformerFactoryMap[v.MetaData.ClusterName] = watcher
			core.HandlerMap[v.MetaData.ClusterName] = handler
			core.DynamicClientMap[v.MetaData.ClusterName] = dyclient
			core.RestConfigMap[v.MetaData.ClusterName] = restConfig
			core.RestMapperMap[v.MetaData.ClusterName] = restMapper
			core.KubectlClientMap[v.MetaData.ClusterName] = kubectlClient
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

// InitClusterCRD 初始化集群资源实例
// FIXME: 目前只有 config.yaml 中配置的集群才会加入，没有动态加入的功能
func (mc *MultiClusterHandler) InitClusterCRD() error {

	for k, v := range mc.RestConfigMap {

		c, err := kubernetes.NewForConfig(v)
		if err != nil {
			klog.Fatal(err)
		}
		version, err := c.Discovery().ServerVersion()
		if err != nil {
			klog.Fatal(err)
		}
		aa := &v1alpha12.MultiCluster{}
		aa.Name = k
		aa.Namespace = "default"
		aa.Spec.Name = k
		aa.Spec.Host = v.Host
		aa.Spec.Platform = version.Platform
		aa.Spec.Version = version.GitVersion

		aa.Spec.IsMaster = "false"
		if k == mc.MasterCluster {
			aa.Spec.IsMaster = "true"
		}

		err = mc.Client.Create(context.Background(), aa)

		if err != nil {
			if errors.IsAlreadyExists(err) {
				return nil
			}
			klog.Fatal(err)
		}
	}

	return nil
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

	// 3. 注册进入序列化表
	err = v1alpha12.SchemeBuilder.AddToScheme(mgr.GetScheme())
	if err != nil {
		mc.Logger.Error(err, "unable add schema")
		return err
	}

	// 4. 赋值 operator 需要的 client EventRecorder
	if mc.Client == nil {
		mc.Client = mgr.GetClient()
	}
	if mc.EventRecorder == nil {
		mc.EventRecorder = mgr.GetEventRecorderFor("multi-cluster-operator1")
	}
	mc.Logger = mgr.GetLogger()

	if err = builder.ControllerManagedBy(mgr).
		For(&v1alpha1.MultiClusterResource{}).
		Complete(mc); err != nil {
		mc.Logger.Error(err, "unable to create manager")
		return err
	}

	cc := NewClusterHandler(mgr.GetClient(), mgr.GetEventRecorderFor("multi-cluster-operator2"))
	if err = builder.ControllerManagedBy(mgr).
		For(&v1alpha12.MultiCluster{}).
		Complete(cc); err != nil {
		mc.Logger.Error(err, "unable to create manager")
		return err
	}

	err = mc.InitClusterCRD()
	if err != nil {
		klog.Fatal(err)
	}

	// 5. 启动controller管理器
	if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
		mc.Logger.Error(err, "unable to start manager")
		return err
	}

	return nil
}
