package config

import (
	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"os"
)

type Config struct {
	Clusters []Cluster `json:"clusters" yaml:"clusters"`
}

func NewConfig() *Config {
	return &Config{}
}

func loadConfigFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		klog.Error("read file error: ", err)
		return nil, err
	}
	return b, nil
}

func BuildConfig(path string) (*Config, error) {
	config := NewConfig()
	if b, err := loadConfigFile(path); b != nil {
		err := yaml.Unmarshal(b, config)
		if err != nil {
			return nil, err
		}
		return config, err
	} else {
		return nil, errors.Wrap(err, "load config file error")
	}
}

// MetaData 集群对象所需的信息
type MetaData struct {
	// ConfigPath kube config文件
	ConfigPath string `json:"configPath" yaml:"configPath"`
	// Insecure 是否跳过证书认证
	Insecure bool `json:"insecure" yaml:"insecure"`
	// ClusterName 集群名
	ClusterName string `json:"clusterName" yaml:"clusterName"`
	// IsMaster 是否为主集群
	IsMaster bool `json:"isMaster" yaml:"isMaster"`
	// Resources 监听的资源对象(用于多集群查询)
	Resources []Resource `json:"resources" yaml:"resources"`
}

type Resource struct {
	// GVR 标示， ex: v1/pods apps/v1/deployments
	RType string `json:"rType" yaml:"rType"`
}

// Cluster 集群对象
type Cluster struct {
	MetaData MetaData `json:"metadata" yaml:"metadata"`
}
