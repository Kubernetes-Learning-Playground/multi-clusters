package config

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"log"
	"os"
)

type Config struct {
	Clusters []Cluster `json:"clusters" yaml:"clusters"`
}

func NewConfig() *Config {
	return &Config{}
}

func loadConfigFile(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return nil
	}
	return b
}

func BuildConfig(path string) (*Config, error) {
	config := NewConfig()
	if b := loadConfigFile(path); b != nil {

		err := yaml.Unmarshal(b, config)
		if err != nil {
			return nil, err
		}
		return config, err
	} else {
		return nil, fmt.Errorf("load config file error")
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
	// isMaster 是否为主集群
	IsMaster bool `json:"isMaster" yaml:"isMaster"`

	Resources []Resource `json:"resources" yaml:"resources"`
}

type Resource struct {
	RType string `json:"rType" yaml:"rType"`
}

// Cluster 集群对象
type Cluster struct {
	MetaData MetaData `json:"metadata" yaml:"metadata"`
}
