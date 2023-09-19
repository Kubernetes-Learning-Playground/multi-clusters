package multi_cluster

import (
	"context"
	"github.com/practice/multi_resource/pkg/caches"
	"github.com/practice/multi_resource/pkg/config"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/klog/v2"
	"strings"
)

// MultiClusterHandler 多集群缓存
type MultiClusterHandler struct {
	informerFactories map[string]dynamicinformer.DynamicSharedInformerFactory
	handler           map[string]*caches.ResourceHandler
}

// NewMultiClusterHandlerFromConfig 输入配置文件目录，返回MultiClusterInformer对象
// 推荐调用者直接使用此方法初始化对象
func NewMultiClusterHandlerFromConfig(path string, db *gorm.DB) (*MultiClusterHandler, error) {

	sysConfig, err := config.BuildConfig(path)
	if err != nil {
		klog.Error("load config error: ", err)
		return nil, err
	}

	return newMultiClusterHandler(sysConfig.Clusters, db)
}

// newMultiClusterHandler
func newMultiClusterHandler(clusters []config.Cluster, db *gorm.DB) (*MultiClusterHandler, error) {

	core := &MultiClusterHandler{
		informerFactories: map[string]dynamicinformer.DynamicSharedInformerFactory{},
		handler:           map[string]*caches.ResourceHandler{},
	}

	for _, v := range clusters {
		if v.MetaData.ConfigPath != "" {
			k8sConfig := config.NewK8sConfig(v.MetaData.ConfigPath, v.MetaData.Insecure)
			watcher := k8sConfig.InitWatchFactory()
			// 用于 GVR GVK 转换
			restMapper := k8sConfig.RestMapper()
			// 初始化回调处理函数
			handler := caches.NewResourceHandler(db, *restMapper, v.MetaData.ClusterName)

			for _, vv := range v.MetaData.Resources {
				gvr := parseGVR(vv.RType)
				watcher.ForResource(gvr).Informer().AddEventHandler(handler)
			}

			core.informerFactories[v.MetaData.ClusterName] = watcher
			core.handler[v.MetaData.ClusterName] = handler
		}
	}

	return core, nil
}

// Start 启动所有 informerFactory 与 work queue
func (mc *MultiClusterHandler) Start(ctx context.Context) {

	for r := range mc.handler {
		klog.Infof("[%s] informer watcher start...", r)
		mc.handler[r].Start(ctx)
		mc.informerFactories[r].Start(wait.NeverStop)
		mc.informerFactories[r].WaitForCacheSync(wait.NeverStop)
	}

}

// parseGVR 解析并指定资源对象 "apps/v1/deployments" "core/v1/resource" "batch/v1/jobs"
func parseGVR(gvr string) schema.GroupVersionResource {
	var group, version, resource string
	gvList := strings.Split(gvr, "/")

	// 防止越界
	if len(gvList) < 2 {
		panic("gvr input error, please input like format apps/v1/deployments or core/v1/resource")
	}

	if len(gvList) < 3 {
		group = ""
		version = gvList[0]
		resource = gvList[1]
	} else {
		if gvList[0] == "core" {
			gvList[0] = ""
		}
		group, version, resource = gvList[0], gvList[1], gvList[2]
	}

	return schema.GroupVersionResource{
		Group: group, Version: version, Resource: resource,
	}
}
