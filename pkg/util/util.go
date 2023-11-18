package util

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"strings"
)

// ParseIntoGvr 解析并指定资源对象GVR，http server 接口使用 "." 作为分割符
// ex: "apps/v1/deployments" "core/v1/pods" "batch/v1/jobs"
// ex："apps.v1.deployments" "v1.pods"
func ParseIntoGvr(gvr, splitString string) schema.GroupVersionResource {
	var group, version, resource string
	gvList := strings.Split(gvr, splitString)

	// 防止越界
	if len(gvList) < 2 {
		panic("gvr input error, please input like format apps/v1/deployments or core/v1/multiclusterresource")
	}

	if len(gvList) == 2 {
		group = ""
		version = gvList[0]
		resource = gvList[1]
	} else {
		if gvList[0] == "core" {
			gvList[0] = ""
		}
		group, version, resource = gvList[0], gvList[1], gvList[2]
	}

	return schema.GroupVersionResource{
		Group: group, Version: version, Resource: resource,
	}
}

// IsNameSpaceScope 是否 namespace 资源
func IsNameSpaceScope(restMapper meta.RESTMapper, gvr schema.GroupVersionResource) bool {
	gvk, err := restMapper.KindFor(gvr)
	if err != nil {
		panic(err)
	}
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvr.Version)
	if err != nil {
		panic(err)
	}
	return string(mapping.Scope.Name()) == "namespace"
}

// contains 是否包含
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func GetDiffString(old []string, new []string) []string {
	// 待删除list
	forDelete := make([]string, 0)
	for _, item := range old {
		// 如果不包含，代表需要删除
		if !contains(new, item) {
			forDelete = append(forDelete, item)
		}
	}
	return forDelete
}

// RemoveItem
// ex: [a,b,c,d,e,f,g] ===> [a,b,d,e,f,g]
func RemoveItem(list []string, item string) []string {
	for i := 0; i < len(list); i++ {
		if list[i] == item {
			list = append(list[:i], list[i+1:]...)
			i--
		}
	}
	return list
}

// GetWd 获取工作目录
func GetWd() string {
	wd := os.Getenv("WORK_DIR")
	if wd == "" {
		wd, _ = os.Getwd()
	}
	return wd
}
