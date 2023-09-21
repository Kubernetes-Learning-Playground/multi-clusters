package multi_cluster_controller

import (
	"github.com/practice/multi_resource/pkg/multi_cluster_controller/helpers"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"

	"log"
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

var (
	DefaultClientSet  kubernetes.Interface
	DefaultRestConfig *rest.Config
	DefaultRestMapper *meta.RESTMapper
)

func getMasterClusterRestMapper() *meta.RESTMapper {

	gr, err := restmapper.GetAPIGroupResources(DefaultClientSet.Discovery())
	if err != nil {
		log.Fatal(err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)
	return &mapper
}

func getMasterClusterClient() kubernetes.Interface {
	client, err := kubernetes.NewForConfig(DefaultRestConfig)
	if err != nil {
		klog.Fatal(err)
	}
	return client
}

// applyCrdToMasterCluster 在主集群中 apply Crd
func (mc *MultiClusterHandler) applyCrdToMasterCluster() {
	if mc.MasterCluster == "" {
		klog.Fatal("masterCluster is empty")
	}

	DefaultRestConfig = mc.RestConfigMap[mc.MasterCluster]
	DefaultClientSet = getMasterClusterClient()
	DefaultRestMapper = getMasterClusterRestMapper()

	jsonBytes, err := yaml.ToJSON([]byte(ResourceCRD))
	if err != nil {
		klog.Fatal(err)
	}

	// 直接 apply crd
	_, err = helpers.K8sApply(jsonBytes, DefaultRestConfig, *DefaultRestMapper)
	if err != nil {
		klog.Fatal(err)
	}

}
