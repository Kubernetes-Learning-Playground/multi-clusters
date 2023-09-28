package config

import (
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
)

// CreateCtlFile 创建命令行工具需要的配置文件
// 默认在 "~/.multi-cluster-operator/config" 中配置
func CreateCtlFile(opt *Options) {
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
	configContent := fmt.Sprintf("server: %v\n", opt.CtlPort)
	err = os.WriteFile(configFilePath, []byte(configContent), 0777)
	if err != nil {
		klog.Errorf("Failed to create config file: %v\n", err)
		return
	}

	klog.Infof("multi-cluster-ctl config file created successfully.")
}
