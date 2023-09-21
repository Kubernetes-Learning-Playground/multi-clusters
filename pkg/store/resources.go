package store

import (
	"github.com/practice/multi_resource/pkg/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"sigs.k8s.io/yaml"
	"time"
)

// Resources 放入表中的模型
type Resources struct {
	// 区分集群
	Cluster string
	// obj Unstructured 结构体，不用于入库
	obj *unstructured.Unstructured `gorm:"-"`
	// objbytes []byte 对象，不用于入库，需要获取 yaml or json 格式使用
	objbytes []byte `gorm:"-"`
	// 主键id
	Id              int    `gorm:"column:id;primaryKey;autoIncrement"`
	Name            string `gorm:"column:name"`
	NameSpace       string `gorm:"column:namespace"`
	ResourceVersion string `gorm:"column:resource_version"`
	// Hash 值
	Hash string `gorm:"column:hash"`
	Uid  string `gorm:"column:uid"`
	// GVR GVK 有关
	Group    string `gorm:"column:group"`
	Version  string `gorm:"column:version"`
	Resource string `gorm:"column:resource"`
	Kind     string `gorm:"column:kind"`
	// owner
	Owner  string `gorm:"column:owner"`
	Object string `gorm:"column:object"`
	// 时间相关，UpdateAt DeleteAt 插入时 不需要赋值
	CreateAt time.Time `gorm:"column:create_at"`
	UpdateAt time.Time `gorm:"column:update_at"`
	DeleteAt time.Time `gorm:"column:delete_at"`
}

func NewResource(obj runtime.Object, restmapper meta.RESTMapper, clusterName string) (*Resources, error) {
	o := &unstructured.Unstructured{}
	b, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, o)
	if err != nil {
		return nil, err
	}

	gvk := o.GroupVersionKind()
	mapping, err := restmapper.RESTMapping(gvk.GroupKind())
	if err != nil {
		return nil, err
	}
	// 赋值
	retObj := &Resources{obj: o, objbytes: b}
	retObj.Cluster = clusterName
	// name namespace
	retObj.Name = o.GetName()
	retObj.NameSpace = o.GetNamespace()
	// gvr gvk
	retObj.Group = gvk.Group
	retObj.Version = gvk.Version
	retObj.Kind = gvk.Kind
	retObj.Resource = mapping.Resource.Resource
	// 版本号
	retObj.ResourceVersion = o.GetResourceVersion()
	retObj.CreateAt = o.GetCreationTimestamp().Time

	retObj.Uid = string(o.GetUID())

	return retObj, nil
}

// prepare  譬如做一些字符串的赋值啊、过滤啊。
func (r *Resources) prepare() {
	// 如果有 OwnerReferences，则设置
	if len(r.obj.GetOwnerReferences()) > 0 {
		r.Owner = string(r.obj.GetOwnerReferences()[0].UID)
	}
	// 获取md5值
	r.Hash = util.HashObject(r.objbytes)
	// 数据库为 Json类型
	//r.Object = string(r.objbytes)
	objJson, err := yaml.YAMLToJSON(r.objbytes)
	if err != nil {
		log.Fatalln(err)
	}
	r.Object = string(objJson)
}

func (r *Resources) Add(db *gorm.DB) error {
	r.prepare()
	// 处理冲突方法：当 uid 发现存在时，只更新 resource_version update_at 字段
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "uid"}},
		DoUpdates: clause.Assignments(
			map[string]interface{}{
				"resource_version": r.ResourceVersion,
				"update_at":        time.Now(),
			}),
	}).Create(r).Error
}

func (r *Resources) Update(db *gorm.DB) error {
	r.prepare()
	r.UpdateAt = time.Now()
	// 当 hash 值不与库中的相等时，才进行更新
	return db.Where("uid=?", r.Uid).Where("hash!=?", r.Hash).Updates(r).Error
}

func (r *Resources) Delete(db *gorm.DB) error {
	return db.Where("uid=?", r.Uid).Delete(r).Error
}

func (*Resources) TableName() string {
	return "resources"
}
