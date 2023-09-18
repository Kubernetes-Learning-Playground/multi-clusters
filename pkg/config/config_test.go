package config

import (
	"fmt"
	"k8s.io/klog/v2"
	"testing"
)

func TestLoadConfig(test *testing.T) {
	// 1. 项目配置
	sysConfig, err := BuildConfig("../../config.yaml")
	if err != nil {
		klog.Error("load config error: ", err)
		return
	}
	SysConfig = sysConfig
	fmt.Println(sysConfig)
	fmt.Println(sysConfig.Clusters[0])
}
