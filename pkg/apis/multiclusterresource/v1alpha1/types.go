package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MultiClusterResource
type MultiClusterResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ResourceSpec   `json:"spec,omitempty"`
	Status            StatusTemplate `json:"status,omitempty"`
}

type ResourceSpec struct {
	// 资源模版
	Template DataTemplate `json:"template,omitempty"`
	// 集群
	/*
		  placement
			clusters:
			   - name: aliyun
			   - name: huawei
	*/
	Placement DataTemplate `json:"placement,omitempty"`
	// 差异化适配
	Customize Customize `json:"customize,omitempty"`
}

type Customize struct {
	Clusters []Cluster `json:"clusters,omitempty"`
}

type Cluster struct {
	Name   string   `json:"name,omitempty"`
	Action []Action `json:"action,omitempty"`
}

type Action struct {
	Type  string        `json:"type,omitempty"`
	Path  string        `json:"path,omitempty"`
	Value []interface{} `json:"value,omitempty"`
	Op    string        `json:"op,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MultiClusterResourceList
type MultiClusterResourceList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MultiClusterResource `json:"items"`
}

// TODO: 需要用到，存储在每个集群中的状态
type StatusTemplate map[string]interface{}

func (in *StatusTemplate) DeepCopyInto(out *StatusTemplate) {
	if in == nil {
		return
	}
	b, err := yaml.Marshal(in)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(b, &out)
	if err != nil {
		return
	}
}

type DataTemplate map[string]interface{}

func (in *DataTemplate) DeepCopyInto(out *DataTemplate) {
	if in == nil {
		return
	}
	b, err := yaml.Marshal(in)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(b, &out)
	if err != nil {
		return
	}

}
