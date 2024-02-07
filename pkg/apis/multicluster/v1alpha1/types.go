package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MultiCluster
type MultiCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClusterSpec   `json:"spec,omitempty"`
	Status            ClusterStatus `json:"status,omitempty"`
}

type ClusterSpec struct {
	// Name 集群名
	Name string `json:"name"`
	// Host 集群地址
	Host string `json:"host"`
	// Version 集群版本
	Version string `json:"version"`
	// Platform 平台版本
	Platform string `json:"platform"`
	// IsMaster 是否为 master 主集群
	IsMaster string `json:"isMaster"`
}

type ClusterStatus struct {
	// Status 集群状态
	Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MultiClusterList
type MultiClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MultiCluster `json:"items"`
}
