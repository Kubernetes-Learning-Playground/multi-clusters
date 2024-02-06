package config

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type K8sConfig struct {
	kubeconfigPath string
	insecure       bool
	restconfig     *rest.Config
	isPatch        bool
}

// NewK8sConfig 创建 K8sConfig 对象，
// 其中： restconfig 对象在 init 时是 nil, 只有在动态时加入才会传入
// isPatch 字段是用于区分是否为初始化加入还是动态加入
func NewK8sConfig(path string, insecure bool, restconfig *rest.Config, isPatch bool) *K8sConfig {
	if isPatch {
		if restconfig != nil {
			restconfig.Insecure = insecure
		}
	}

	return &K8sConfig{
		kubeconfigPath: path,
		insecure:       insecure,
		restconfig:     restconfig,
		isPatch:        isPatch,
	}
}

func (kc *K8sConfig) k8sRestConfigDefaultOrDie(insecure bool) *rest.Config {

	config, err := clientcmd.BuildConfigFromFlags("", kc.kubeconfigPath)
	if err != nil {
		klog.Fatal(err)
	}
	config.Insecure = insecure
	return config
}

// initDynamicClientOrDie 初始化 DynamicClient
func (kc *K8sConfig) initDynamicClientOrDie() dynamic.Interface {
	client, err := dynamic.NewForConfig(kc.k8sRestConfigDefaultOrDie(kc.insecure))
	if err != nil {
		klog.Fatal(err)
	}
	return client
}

// initClient 初始化 clientSet
func (kc *K8sConfig) initClientOrDie() *kubernetes.Clientset {
	var err error
	var c *kubernetes.Clientset
	if kc.isPatch {
		c, err = kubernetes.NewForConfig(kc.restconfig)
	} else {
		c, err = kubernetes.NewForConfig(kc.k8sRestConfigDefaultOrDie(kc.insecure))
	}

	if err != nil {
		klog.Fatal(err)
	}
	return c
}

// NewRestMapperOrDie 获取 api group multiclusterresource
func (kc *K8sConfig) NewRestMapperOrDie() *meta.RESTMapper {
	gr, err := restmapper.GetAPIGroupResources(kc.initClientOrDie().Discovery())
	if err != nil {
		klog.Fatal(err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)
	return &mapper
}

// InitWatchFactoryAndRestConfig 初始化 dynamic client informerFactory, restConfig
func (kc *K8sConfig) InitWatchFactoryAndRestConfig() (dynamicinformer.DynamicSharedInformerFactory, dynamic.Interface, *rest.Config) {
	dynClient := kc.initDynamicClientOrDie()
	return dynamicinformer.NewDynamicSharedInformerFactory(dynClient, 0), dynClient, kc.k8sRestConfigDefaultOrDie(kc.insecure)
}

func (kc *K8sConfig) PatchWatchFactoryAndRestConfig() (dynamicinformer.DynamicSharedInformerFactory, dynamic.Interface, *rest.Config) {

	dynClient, err := dynamic.NewForConfig(kc.restconfig)
	if err != nil {
		klog.Fatal(err)
	}
	return dynamicinformer.NewDynamicSharedInformerFactory(dynClient, 0), dynClient, kc.restconfig
}
