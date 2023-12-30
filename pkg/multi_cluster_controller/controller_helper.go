package multi_cluster_controller

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/practice/multi_resource/pkg/apis/multiclusterresource/v1alpha1"
	"github.com/practice/multi_resource/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"strings"

	"sort"
)

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

type JSONPatch struct {
	Path  string      `json:"path"`
	Op    string      `json:"op"`
	Value interface{} `json:"value,omitempty"`
}

type ParseRes struct {
	Cluster string
	Action  []*JSONPatch
}

// getCustomizeClusters 从 crd 获取差异化信息
// FIXME: 注意 这种处理切片很容易 panic，可能会有没考虑到的地方
func (mc *MultiClusterHandler) getCustomizeClusters(res *v1alpha1.MultiClusterResource) []*ParseRes {
	resultList := make([]*ParseRes, 0)

	if len(res.Spec.Customize.Clusters) == 0 {
		return resultList
	}

	for _, v := range res.Spec.Customize.Clusters {
		jsonPatchList := &ParseRes{
			Cluster: v.Name,
			Action:  make([]*JSONPatch, 0),
		}
		for _, action := range v.Action {
			jsonPatch := &JSONPatch{
				Path: action.Path,
				Op:   action.Op,
			}

			patchMap := make(map[string]interface{})
			var patchInterface interface{}
			var is bool
			for _, value := range action.Value {
				// 判断是否是 string 类型，如果是会有两种情况
				// 1. 输入是 类似 "name=redis" or "image=xxx"
				// 所以需要分割存入 map
				// 2. 直接是普通 string
				if dd, ok := value.(string); ok {
					caa := strings.Split(dd, "=")

					if len(caa) > 1 {
						is = true
						patchMap[caa[0]] = caa[1]
					}
				}
				// 接受其他类型 ex: replicas 需要的是 int32
				patchInterface = value
			}

			// 区分赋值 patchMap or patchInterface
			if is {
				jsonPatch.Value = patchMap
			} else {
				jsonPatch.Value = patchInterface
			}

			jsonPatchList.Action = append(jsonPatchList.Action, jsonPatch)
		}
		resultList = append(resultList, jsonPatchList)
	}

	return resultList
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
func (mc *MultiClusterHandler) resourceDeleteBySlice(ctx context.Context, res *v1alpha1.MultiClusterResource, clusters []string) error {
	tpl := res.Spec.Template
	obj := &unstructured.Unstructured{}
	obj.SetUnstructuredContent(tpl)

	b, err := obj.MarshalJSON()
	if err != nil {
		return err
	}
	// 遍历获取 restConfig 并删除
	for _, c := range clusters {
		if _, ok := mc.RestConfigMap[c]; ok {
			//err := helpers.K8sDelete(b, cfg, *mc.RestMapperMap[c])
			//if err != nil {
			//	return err
			//}
			err := mc.KubectlClientMap[c].Delete(ctx, b, false)
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
		if _, ok := mc.RestConfigMap[c]; ok {
			err = mc.KubectlClientMap[c].Delete(context.Background(), b, true)
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
	// TODO: OwnerReference 字段沒有設置
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

		err = mc.KubectlClientMap[mc.MasterCluster].Apply(context.Background(), b)
		if err != nil {
			return err
		}
	} else {
		for _, c := range clusters {
			if _, ok := mc.RestConfigMap[c]; ok {
				err = mc.KubectlClientMap[c].Apply(context.Background(), b)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func GetGVR(obj *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	accessor := meta.NewAccessor()
	apiVersion, err := accessor.APIVersion(obj)
	r := schema.GroupVersionResource{}
	if err != nil {
		return r, fmt.Errorf("failed to get API version: %v", err)
	}

	kind, err := accessor.Kind(obj)
	if err != nil {
		return r, fmt.Errorf("failed to get kind: %v", err)
	}
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return r, fmt.Errorf("failed to parse GroupVersion: %v", err)
	}
	// FIXME: 目前使用 转小写 + s 的方式 mock
	resource := strings.ToLower(kind) + "s"
	gvr := gv.WithResource(resource)
	return gvr, nil
}

func (mc *MultiClusterHandler) resourcePatch(res *v1alpha1.MultiClusterResource) error {
	tpl := res.Spec.Template
	obj := &unstructured.Unstructured{}
	obj.SetUnstructuredContent(tpl)
	gvr, err := GetGVR(obj)
	if err != nil {
		return err
	}

	// 处理 Placement
	clusters := mc.getCustomizeClusters(res)

	// 区分需要对哪些集群进行 patch
	if len(clusters) == 0 {
		return nil

	} else {
		for _, c := range clusters {

			if dyclient, ok := mc.DynamicClientMap[c.Cluster]; ok {
				b, err := json.Marshal(c.Action)
				if err != nil {
					return errors.Wrap(err, "patch action marshal error")
				}
				_, err = dyclient.Resource(gvr).Namespace(obj.GetNamespace()).
					Patch(context.Background(), obj.GetName(), types.JSONPatchType, b, metav1.PatchOptions{})
				if err != nil {
					return errors.Wrap(err, "patch to api-server error")
				}
			}
		}
	}
	return nil
}
