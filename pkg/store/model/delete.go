package model

import "gorm.io/gorm"

func DeleteResourcesByClusterName(db *gorm.DB, clusterName string) error {
	rr := Resources{Cluster: clusterName}
	return db.Where("cluster=?", clusterName).Delete(rr).Error
}
