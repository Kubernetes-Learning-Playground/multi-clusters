package config

import (
	"fmt"
	"github.com/myoperator/multiclusteroperator/pkg/options"
	"io"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
)

// CreateCtlFile 创建命令行工具需要的配置文件
// 默认在 "~/.multi-cluster-operator/config" 中配置
// FIXME: 容器里创建没用，考虑废弃。。。。
func CreateCtlFile(opt *options.ServerOptions, masterClusterKubeConfigPath string) {
	// 获取用户的 Home 目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		klog.Errorf("Failed to get user's home directory: %v\n", err)
		return
	}

	// 创建目录
	dirPath := filepath.Join(homeDir, ".multi-cluster-operator")
	err = os.MkdirAll(dirPath, 0777)
	if err != nil {
		klog.Errorf("Failed to create directory: %v\n", err)
		return
	}

	// 创建配置文件
	configFilePath := filepath.Join(dirPath, "config")
	configContent := fmt.Sprintf("serverIP: %v\nserverPort: %v\nmasterClusterKubeConfigPath: %v", "localhost", opt.CtlPort, masterClusterKubeConfigPath)

	// 创建或覆盖文件
	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		klog.Fatalf("Error creating or truncating file: %s\n", err)
	}
	defer file.Close()

	_, err = io.WriteString(file, configContent)
	if err != nil {
		klog.Fatalf("Failed to create config file: %v\n", err)
		return
	}

	klog.Infof("multi-cluster-ctl config file created successfully.")
}
