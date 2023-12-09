package model

import "time"

// Cluster 放入表中的模型
type Cluster struct {
	Id int `gorm:"column:id;primaryKey;autoIncrement"`
	// 集群名称
	Name string `gorm:"column:name"`
	// 是否为 master 主集群
	IsMaster string `gorm:"column:isMaster"`
	// 状态
	Status string `gorm:"column:status"`
	// 时间相关，UpdateAt DeleteAt 插入时 不需要赋值
	CreateAt time.Time `gorm:"column:create_at"`
}

func (*Cluster) TableName() string {
	return "clusters"
}

func NewCluster(name string, isMaster bool) *Cluster {
	c := &Cluster{
		Name:     name,
		CreateAt: time.Now(),
		Status:   "Running",
	}
	if isMaster {
		c.IsMaster = "true"
	} else {
		c.IsMaster = "false"
	}

	return c
}
