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
	"github.com/myoperator/multiclusteroperator/pkg/options/mysql"
	"github.com/myoperator/multiclusteroperator/pkg/store/model"
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
	"sync"
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
	KubectlClientMap   map[string]*kubectl_client.KubectlManager
	RestMapperMap      map[string]*meta.RESTMapper
	InformerFactoryMap map[string]dynamicinformer.DynamicSharedInformerFactory
	HandlerMap         map[string]*caches.ResourceHandler
	// operator 控制器 client
	client.Client
	// 事件发送器
	EventRecorder record.EventRecorder

	logr.Logger
}

// GlobalMultiClusterHandler 全局变量，用于存放多集群的客户端实例
var GlobalMultiClusterHandler *MultiClusterHandler

func init() {
	GlobalMultiClusterHandler = &MultiClusterHandler{
		RestConfigMap:      map[string]*rest.Config{},
		RestMapperMap:      map[string]*meta.RESTMapper{},
		DynamicClientMap:   map[string]dynamic.Interface{},
		InformerFactoryMap: map[string]dynamicinformer.DynamicSharedInformerFactory{},
		HandlerMap:         map[string]*caches.ResourceHandler{},
		KubectlClientMap:   map[string]*kubectl_client.KubectlManager{},
	}
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

// AddMultiClusterHandler 加入集群
func AddMultiClusterHandler(cluster *config.Cluster) (err error) {
	defer func() {
		if err != nil {
			klog.Errorf("delete cluster error: ", err)
		}
	}()
	k8sConfig := config.NewK8sConfig(cluster.MetaData.ConfigPath, cluster.MetaData.Insecure, cluster.MetaData.RestConfig, true)
	watcher, dyclient, restConfig := k8sConfig.PatchWatchFactoryAndRestConfig()
	// 用于 GVR GVK 转换
	restMapper := k8sConfig.NewRestMapperOrDie()
	// 初始化回调处理函数

	handler := caches.NewResourceHandler(mysql.GlobalDB, *restMapper, cluster.MetaData.ClusterName)
	kubectlClient := kubectl_client.NewKubectlManagerOrDie(restConfig)

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

			gvr, err := util.ParseIntoGvr(fmt.Sprintf("%v/%v/%v", groupVersion.Group, groupVersion.Version, apiResource.Name), "/")
			if err != nil {
				continue
			}
			_, err = watcher.ForResource(gvr).Informer().AddEventHandler(handler)
			if err != nil {
				continue
			}
		}
	}

	klog.Infof("[%s] informer watcher start..\n", cluster.MetaData.ClusterName)
	handler.Start(context.Background())
	watcher.Start(wait.NeverStop)
	watcher.WaitForCacheSync(wait.NeverStop)
	err = GlobalMultiClusterHandler.AddClusterCRD(cluster)
	if err != nil {
		return err
	}

	mc := sync.Mutex{}
	mc.Lock()
	defer mc.Unlock()

	GlobalMultiClusterHandler.InformerFactoryMap[cluster.MetaData.ClusterName] = watcher
	GlobalMultiClusterHandler.HandlerMap[cluster.MetaData.ClusterName] = handler
	GlobalMultiClusterHandler.DynamicClientMap[cluster.MetaData.ClusterName] = dyclient
	GlobalMultiClusterHandler.RestConfigMap[cluster.MetaData.ClusterName] = restConfig
	GlobalMultiClusterHandler.RestMapperMap[cluster.MetaData.ClusterName] = restMapper
	GlobalMultiClusterHandler.KubectlClientMap[cluster.MetaData.ClusterName] = kubectlClient
	return nil
}

func DeleteMultiClusterHandlerByClusterName(clusterName string) (err error) {
	defer func() {
		if err != nil {
			klog.Errorf("delete cluster error: ", err)
		}
	}()
	// 限制：不能删除主集群
	if clusterName == GlobalMultiClusterHandler.MasterCluster {
		klog.Errorf("cannot delete master cluster")
		return errors.NewBadRequest("cannot delete master cluster")
	}

	// 1. 把 db 中有关的数据删除
	err = model.DeleteResourcesByClusterName(mysql.GlobalDB, clusterName)
	if err != nil {
		return err
	}
	// 2. 删除 multi-cluster 资源对象
	err = GlobalMultiClusterHandler.DeleteClusterCRD(clusterName)
	if err != nil {
		return err
	}

	// 3. 由 map 删除
	// 使用锁防止并发
	mc := sync.Mutex{}
	mc.Lock()
	defer mc.Unlock()
	delete(GlobalMultiClusterHandler.InformerFactoryMap, clusterName)
	delete(GlobalMultiClusterHandler.DynamicClientMap, clusterName)
	delete(GlobalMultiClusterHandler.RestConfigMap, clusterName)
	delete(GlobalMultiClusterHandler.RestMapperMap, clusterName)
	delete(GlobalMultiClusterHandler.KubectlClientMap, clusterName)
	return nil
}

// newMultiClusterHandler 初始化各集群需要的资源
func newMultiClusterHandler(clusters []config.Cluster, db *gorm.DB) (*MultiClusterHandler, error) {

	if len(clusters) == 0 {
		panic("empty cluster...")
	}

	core := GlobalMultiClusterHandler

	// 遍历
	for _, v := range clusters {
		// 如果有主集群
		if v.MetaData.IsMaster {
			core.MasterCluster = v.MetaData.ClusterName
			core.MasterClusterKubeConfigPath = v.MetaData.ConfigPath
		}
		// 处理需要的初始化
		if v.MetaData.ConfigPath != "" {
			k8sConfig := config.NewK8sConfig(v.MetaData.ConfigPath, v.MetaData.Insecure, nil, false)
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

					//klog.Infof("GVR: %v/%v/%v\n", groupVersion.Group, groupVersion.Version, apiResource.Name)

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

					gvr, err := util.ParseIntoGvr(fmt.Sprintf("%v/%v/%v", groupVersion.Group, groupVersion.Version, apiResource.Name), "/")
					if err != nil {
						continue
					}
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

// initClusterCRD 初始化集群资源实例
func (mc *MultiClusterHandler) initClusterCRD() error {

	for clusterName, v := range mc.RestConfigMap {
		// 获取 version 版本，有多次调用的
		c, err := kubernetes.NewForConfig(v)
		if err != nil {
			klog.Fatal(err)
		}
		version, err := c.Discovery().ServerVersion()
		if err != nil {
			klog.Fatal(err)
		}
		multiCluster := &v1alpha12.MultiCluster{}
		multiCluster.Name = clusterName
		multiCluster.Namespace = "default"
		multiCluster.Spec.Name = clusterName
		multiCluster.Spec.Host = v.Host
		multiCluster.Spec.Platform = version.Platform
		multiCluster.Spec.Version = version.GitVersion
		multiCluster.Spec.IsMaster = "false"
		if clusterName == mc.MasterCluster {
			multiCluster.Spec.IsMaster = "true"
		}
		err = mc.Client.Create(context.Background(), multiCluster)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				return nil
			}
			klog.Fatal(err)
		}
	}
	return nil
}

// AddClusterCRD 加入集群实例
func (mc *MultiClusterHandler) AddClusterCRD(cluster *config.Cluster) error {

	c, err := kubernetes.NewForConfig(cluster.MetaData.RestConfig)
	if err != nil {
		klog.Fatal(err)
	}
	version, err := c.Discovery().ServerVersion()
	if err != nil {
		klog.Fatal(err)
	}
	multiCluster := &v1alpha12.MultiCluster{}
	multiCluster.Name = cluster.MetaData.ClusterName
	multiCluster.Namespace = "default"
	multiCluster.Spec.Name = cluster.MetaData.ClusterName
	multiCluster.Spec.Host = cluster.MetaData.RestConfig.Host
	multiCluster.Spec.Platform = version.Platform
	multiCluster.Spec.Version = version.GitVersion

	multiCluster.Spec.IsMaster = "false"

	err = mc.Client.Create(context.Background(), multiCluster)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		klog.Fatal(err)
	}

	return nil
}

// DeleteClusterCRD 加入集群实例
func (mc *MultiClusterHandler) DeleteClusterCRD(clusterName string) error {
	multiCluster := &v1alpha12.MultiCluster{}
	multiCluster.Name = clusterName
	multiCluster.Namespace = "default"
	err := mc.Client.Delete(context.Background(), multiCluster)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
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
	// 创建 cluster 对象
	err = mc.initClusterCRD()
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
