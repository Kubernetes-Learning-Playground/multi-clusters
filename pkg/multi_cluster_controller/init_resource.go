package multi_cluster_controller

import (
	"context"
	v1alpha12 "github.com/myoperator/multiclusteroperator/pkg/apis/multicluster/v1alpha1"
	"github.com/myoperator/multiclusteroperator/pkg/apis/multiclusterresource/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const ResourceCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # 名字必需与下面的 spec group字段匹配，并且格式为 '<名称的复数形式>.<组名>'
  name: multiclusterresources.mulitcluster.practice.com
  labels:
    version: "0.1"
spec:
  group: mulitcluster.practice.com
  versions:
    - name: v1alpha1
      # 是否有效
      served: true
      #是否是存储版本
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          #没有任何内容会被修剪，哪怕不被识别
          x-kubernetes-preserve-unknown-fields: true
      subresources:
        status: {}
  names:
    # 复数名
    plural: multiclusterresources
    # 单数名
    singular: multiclusterresource
    kind: MultiClusterResource
    listKind: MultiClusterResourceList
    # kind的简称
    shortNames:
      - rr
  scope: Namespaced
`

const ClusterCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # 名字必需与下面的 spec group字段匹配，并且格式为 '<名称的复数形式>.<组名>'
  name: multiclusters.mulitcluster.practice.com
  labels:
    version: "0.1"
spec:
  group: mulitcluster.practice.com
  versions:
    - name: v1alpha1
      # 是否有效
      served: true
      #是否是存储版本
      storage: true
      additionalPrinterColumns:
        - name: Version
          type: string
          jsonPath: .spec.version
        - name: Host
          type: string
          jsonPath: .spec.host
        - name: Platform
          type: string
          jsonPath: .spec.platform
        - name: IsMaster
          type: string
          jsonPath: .spec.isMaster  
        - name: Status
          type: string
          jsonPath: .status.status
        - name: Age
          type: date
          jsonPath: .metadata.creationTimestamp
      schema:
        openAPIV3Schema:
          type: object
          #没有任何内容会被修剪，哪怕不被识别
          x-kubernetes-preserve-unknown-fields: true
      subresources:
        status: {}
  names:
    # 复数名
    plural: multiclusters
    # 单数名
    singular: multicluster
    kind: MultiCluster
    listKind: MultiClusterList
    # kind的简称
    shortNames:
      - cl
  scope: Namespaced
`

var (
	DefaultClientSet  kubernetes.Interface
	DefaultRestConfig *rest.Config
)

// CRDsInstalled checks if the CRDs are installed or not
func checkCRDsInstalled(discovery discovery.DiscoveryInterface) bool {
	gvs := []schema.GroupVersionKind{
		v1alpha1.SchemeGroupVersion.WithKind(v1alpha1.MultiClusterResourceKind),
		v1alpha1.SchemeGroupVersion.WithKind(v1alpha12.MultiClusterKind),
	}

	for _, gv := range gvs {
		if !isCRDInstalled(discovery, gv) {
			return false
		}
	}

	return true
}

func isCRDInstalled(discovery discovery.DiscoveryInterface, gvk schema.GroupVersionKind) bool {
	crdList, err := discovery.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		klog.ErrorS(err, "resource not found", "resource", gvk)
		return false
	}

	for _, crd := range crdList.APIResources {
		if crd.Kind == gvk.Kind {
			klog.InfoS("resource CRD not found", "resource", crd.Kind)
			return true
		}
	}
	return false
}

func getMasterClusterClient() kubernetes.Interface {
	client, err := kubernetes.NewForConfig(DefaultRestConfig)
	if err != nil {
		klog.Fatal(err)
	}
	return client
}

// applyCrdToMasterCluster 在主集群中 apply CRD
// 目前将两个 CRD 全都内置到主集群中
func (mc *MultiClusterHandler) applyCrdToMasterClusterOrDie() {
	if mc.MasterCluster == "" {
		klog.Fatal("masterCluster is empty")
	}

	DefaultRestConfig = mc.RestConfigMap[mc.MasterCluster]
	DefaultClientSet = getMasterClusterClient()

	// check crd resource
	if checkCRDsInstalled(DefaultClientSet.Discovery()) {
		return
	}

	// apply 第一个
	jsonBytes, err := yaml.ToJSON([]byte(ResourceCRD))
	if err != nil {
		klog.Fatal(err)
	}

	err = mc.KubectlClientMap[mc.MasterCluster].Apply(context.Background(), jsonBytes)
	if err != nil {
		klog.Fatal(err)
	}

	// apply 第二个
	cjsonBytes, err := yaml.ToJSON([]byte(ClusterCRD))
	if err != nil {
		klog.Fatal(err)
	}

	err = mc.KubectlClientMap[mc.MasterCluster].Apply(context.Background(), cjsonBytes)
	if err != nil {
		klog.Fatal(err)
	}

}
