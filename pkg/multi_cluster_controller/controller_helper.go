package multi_cluster_controller

import (
	"github.com/practice/multi_resource/pkg/apis/resource/v1alpha1"
	"github.com/practice/multi_resource/pkg/multi_cluster_controller/helpers"
	"github.com/practice/multi_resource/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sort"
)

type Placement struct {
	Clusters []string `json:"clusters"`
}

// getPlacementClusters 从 crd 获取需要下发到哪些集群
func (mc *MultiClusterHandler) getPlacementClusters(res *v1alpha1.MultiClusterResource) []string {
	clustersList := make([]string, 0)

	if res.Spec.Placement == nil {
		return clustersList
	}

	// 一个个解析
	if clusters, ok := res.Spec.Placement["clusters"]; ok {
		if clusterList, ok := clusters.([]interface{}); ok {
			for _, clusterItem := range clusterList {
				if clusterObj, ok := clusterItem.(map[string]interface{}); ok {
					if clusterName, ok := clusterObj["name"]; ok {
						if _, ok := mc.RestConfigMap[clusterName.(string)]; ok {
							clustersList = append(clustersList, clusterName.(string))
						}
					}
				}

			}
		}
	}
	return clustersList
}

// setResourceFinalizer 比对进入调协的 crd 的 Finalizers 字段与 Placement 字段是否一致
// 返回值：1. 需要被删除的集群名 list 2. 最新的 Finalizers list (比较后最新的) 3. 是否有变化
func (mc *MultiClusterHandler) setResourceFinalizer(res *v1alpha1.MultiClusterResource) ([]string, []string, bool) {
	md5Old := ""
	if res.Finalizers == nil {
		res.Finalizers = make([]string, 0)
	} else {
		md5Old = util.Md5slice(res.Finalizers)
	}

	// 在 Finalizers 中记录 集群列表
	cluster := mc.getPlacementClusters(res)

	// 如果没有设置 Placement ，默认只有主集群
	if len(cluster) == 0 {
		cluster = append(cluster, mc.MasterCluster)
	}

	fz := make([]string, 0)
	fz = append(fz, cluster...)
	// 先排序，再 md5 ，保证一致性
	sort.StringSlice(fz).Sort()
	md5New := util.Md5slice(fz)

	// 比校 md5Old(Finalizers 字段) 与 md5New(Placement 字段)
	// 如果有变化，找出差异，且返回变化后结果
	if md5Old == md5New {
		// 如果没变化，代表不需要 update
		return nil, nil, false
	} else {
		old := res.Finalizers
		forDelete := util.GetDiffString(old, fz)
		return forDelete, fz, true
	}
}

// resourceDeleteBySlice 由传入的 list 删除资源对象，用于更新操作时使用
func (mc *MultiClusterHandler) resourceDeleteBySlice(res *v1alpha1.MultiClusterResource, clusters []string) error {
	tpl := res.Spec.Template
	obj := &unstructured.Unstructured{}
	obj.SetUnstructuredContent(tpl)

	b, err := obj.MarshalJSON()
	if err != nil {
		return err
	}
	// 遍历获取 restConfig 并删除
	for _, c := range clusters {
		if cfg, ok := mc.RestConfigMap[c]; ok {
			err := helpers.K8sDelete(b, cfg, *mc.RestMapperMap[c])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// resourceDelete 删除集群内资源时使用
func (mc *MultiClusterHandler) resourceDelete(res *v1alpha1.MultiClusterResource) error {
	tpl := res.Spec.Template
	obj := &unstructured.Unstructured{}
	obj.SetUnstructuredContent(tpl)

	b, err := obj.MarshalJSON()
	if err != nil {
		return err
	}

	clusters := res.Finalizers
	deletedClusters := make([]string, 0)

	for _, c := range clusters {
		if cfg, ok := mc.RestConfigMap[c]; ok {
			err = helpers.K8sDelete(b, cfg, *mc.RestMapperMap[c])
			if err != nil {
				return err
			}
			deletedClusters = append(deletedClusters, c)
		}
	}

	// 遍历被删除的集群 从 Finalizers 里去除
	for _, c := range deletedClusters {
		res.Finalizers = util.RemoveItem(res.Finalizers, c)
	}
	return nil
}

// resourceApply 资源创建
func (mc *MultiClusterHandler) resourceApply(res *v1alpha1.MultiClusterResource) error {
	tpl := res.Spec.Template
	obj := &unstructured.Unstructured{}
	obj.SetUnstructuredContent(tpl)

	b, err := obj.MarshalJSON()
	if err != nil {
		return err
	}

	// 处理 Placement
	clusters := mc.getPlacementClusters(res)

	// 区分需要对哪些集群进行 apply
	if len(clusters) == 0 {
		_, err = helpers.K8sApply(b, DefaultRestConfig, *DefaultRestMapper)
		if err != nil {
			return err
		}
	} else {
		for _, c := range clusters {
			if cfg, ok := mc.RestConfigMap[c]; ok {
				_, err = helpers.K8sApply(b, cfg, *mc.RestMapperMap[c])
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
