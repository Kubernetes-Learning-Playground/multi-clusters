package service

import (
	"github.com/myoperator/multiclusteroperator/pkg/store/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

type WrapWithCluster struct {
	runtime.Object
	ClusterName string `json:"clusterName"`
}

// ListWrapWithCluster 从数据库获取查询结果
func (list *ListService) ListWrapWithCluster(name, namespace, cluster string, labels map[string]string, gvr schema.GroupVersionResource,
	limit int) ([]WrapWithCluster, error) {
	ret := make([]model.Resources, 0)

	// gvr 一定会传入
	db := list.DB.Model(&model.Resources{}).
		Where("`group`=?", gvr.Group).
		Where("version=?", gvr.Version).
		Where("resource=?", gvr.Resource)
	//Where("object->'$.metadata.labels.app'=?", "test")

	// 其他查询字段自由传入

	if cluster != "" {
		db = db.Where("cluster=?", cluster)
	}

	if name != "" {
		db = db.Where("name=?", name)
	}

	if namespace != "" {
		db = db.Where("namespace=?", namespace)
	}

	if limit != 0 {
		db = db.Limit(limit)
	}

	err := db.Order("create_at desc").Find(&ret).Error
	if err != nil {
		return nil, err
	}

	objList := make([]WrapWithCluster, len(ret))
	for i, res := range ret {
		obj := &unstructured.Unstructured{}
		if err = obj.UnmarshalJSON([]byte(res.Object)); err != nil {
			klog.Errorf("unmarshal json from db error: %s\n", err)
		} else {
			objList[i].Object = obj
			objList[i].ClusterName = res.Cluster
		}
	}

	return objList, err
}
