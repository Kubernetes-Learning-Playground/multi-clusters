package store

import (
	"crypto/md5"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"sigs.k8s.io/yaml"
	"time"
)

// hashObject 序列化内容进行md5
func hashObject(data []byte) string {
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)

}

// Resources 放入表中的模型
type Resources struct {
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

	//-- uid是唯一的 。owner 不一定有
	Owner string `gorm:"column:owner"`
	//-- end uid and owner
	Object string `gorm:"column:object"`

	// ----时间相关
	CreateAt time.Time `gorm:"column:create_at"`
	UpdateAt time.Time `gorm:"column:update_at"`
	DeleteAt time.Time `gorm:"column:delete_at"`
	//--UpdateAt DeleteAt 插入时 不需要赋值
}

func NewResource(obj runtime.Object, restmapper meta.RESTMapper) (*Resources, error) {
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
	retObj.Name = o.GetName()
	retObj.NameSpace = o.GetNamespace()
	retObj.Group = gvk.Group
	retObj.Version = gvk.Version
	retObj.Kind = gvk.Kind
	retObj.Resource = mapping.Resource.Resource
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
	r.Hash = hashObject(r.objbytes)
	// 这个部分改掉了  --数据库字段改成了JSON
	//r.Object = string(r.objbytes)
	objJson, err := yaml.YAMLToJSON(r.objbytes)
	if err != nil {
		log.Fatalln(err)
	}
	r.Object = string(objJson)
}

// FIXME: 这里不应该直接由 informer 触发，应该先放入一个工作队列，由队列触发

func (r *Resources) Add(db *gorm.DB) error {
	r.prepare()
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
	return db.Where("uid=?", r.Uid).Where("hash!=?", r.Hash).Updates(r).Error
}

func (*Resources) TableName() string {
	return "resources"
}
