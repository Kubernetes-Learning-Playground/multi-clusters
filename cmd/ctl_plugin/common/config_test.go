package common

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFile(t *testing.T) {
	// 获取用户的 Home 目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get user's home directory: %v\n", err)
		return
	}

	// 创建目录
	dirPath := filepath.Join(homeDir, ".multi-cluster-operator")
	err = os.MkdirAll(dirPath, 0777)
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		return
	}

	// 创建配置文件
	configFilePath := filepath.Join(dirPath, "config")
	configContent := "server: 8888\n"
	err = os.WriteFile(configFilePath, []byte(configContent), 0777)
	if err != nil {
		fmt.Printf("Failed to create config file: %v\n", err)
		return
	}

	fmt.Println("Config file created successfully.")
}
