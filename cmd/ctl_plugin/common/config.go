package common

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"k8s.io/klog/v2"
	"log"
	"os"
)

// CtlConfig 命令行配置文件
type CtlConfig struct {
	// Server 端口
	ServerIP   string `yaml:"serverIP"`
	ServerPort string `yaml:"serverPort"`
	Token      string `yaml:"token"`
}

// LoadConfigFile 读取配置文件,模仿kubectl，默认在~/.multi-cluster-operator/config
func LoadConfigFile() *CtlConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}
	configFile := fmt.Sprintf("%s/.multi-cluster-operator/config", home)
	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		klog.Fatal("config file not found")
	}

	// 接配置文件
	cfg := &CtlConfig{}
	err = yaml.Unmarshal(MustLoadFile(configFile), cfg)
	if err != nil {
		log.Fatalln(err)
	}
	return cfg
}

// MustLoadFile 如果读不到file，就panic
func MustLoadFile(path string) []byte {
	b, err := LoadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}

// LoadFile 加载指定目录的文件, 全部取出内容
func LoadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, err
}
