package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/practice/multi_resource/pkg/store"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"strconv"
	"strings"
)

type ListService struct {
	DB *gorm.DB
}

// List 从数据库获取查询结果
func (list *ListService) List(name, namespace, cluster string, labels map[string]string, gvr schema.GroupVersionResource,
	limit int) ([]runtime.Object, error) {
	ret := make([]store.Resources, 0)

	// gvr 一定会传入
	db := list.DB.Model(&store.Resources{}).
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

	// FIXME: labels支持有问题
	//if labels != nil {
	//	for k, v := range labels {
	//		db = db.Where(fmt.Sprintf("object->'$.metadata.labels.%s'=?", k), v)
	//	}
	//}

	if limit != 0 {
		db = db.Limit(limit)
	}

	err := db.Order("create_at desc").Find(&ret).Error
	if err != nil {
		return nil, err
	}
	// 列出 runtime.Object
	objList := make([]runtime.Object, len(ret))
	for i, res := range ret {
		obj := &unstructured.Unstructured{}
		if err = obj.UnmarshalJSON([]byte(res.Object)); err != nil {
			log.Println(err)
		} else {
			objList[i] = obj
		}
	}

	return objList, err
}

type ResourceController struct {
	ListService *ListService `inject:"-"`
}

// List 对外暴露接口
func (r *ResourceController) List(c *gin.Context) {
	// 接口传入参数
	gvrParam := c.Query("gvr")
	name := c.DefaultQuery("name", "")
	ns := c.DefaultQuery("namespace", "")
	limit := c.DefaultQuery("limit", "10")
	cluster := c.DefaultQuery("cluster", "")
	var labels map[string]string // 默认是nil
	if labelsQuery, ok := c.GetQueryMap("labels"); ok {
		labels = labelsQuery
	}

	// 解析 gvr
	gvr := parseIntoGvr(gvrParam)
	if gvr.Empty() {
		panic("error gvr")
	}
	ll, _ := strconv.Atoi(limit)
	// 获取列表
	list, err := r.ListService.List(name, ns, cluster, labels, gvr, ll)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}

	c.JSON(200, list)
	return
}

// parseIntoGvr 支持类型："apps.v1.deployments" "v1.pods"
func parseIntoGvr(gvr string) schema.GroupVersionResource {
	list := strings.Split(gvr, ".")
	ret := schema.GroupVersionResource{}
	if len(list) < 2 {
		panic("gvr input error, please input like format apps.v1.deployments or v1.resource")
	}
	// 区分
	if len(list) == 2 {
		ret.Version, ret.Resource = list[0], list[1]
	} else if len(list) > 2 {
		lastIndex := len(list) - 1
		ret.Version, ret.Resource = list[lastIndex-1], list[lastIndex]
		ret.Group = strings.Join(list[0:lastIndex-1], ".")
	}
	return ret
}

// isNameSpaceScope 是否 namespace 资源
func isNameSpaceScope(restMapper meta.RESTMapper, gvr schema.GroupVersionResource) bool {
	gvk, err := restMapper.KindFor(gvr)
	if err != nil {
		panic(err)
	}
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvr.Version)
	if err != nil {
		panic(err)
	}
	return string(mapping.Scope.Name()) == "namespace"
}
