package httpserver

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

// parseIntoGvr 支持类型："apps.v1.deployments" "v1.pods"
func parseIntoGvr(gvr string) schema.GroupVersionResource {
	list := strings.Split(gvr, ".")
	ret := schema.GroupVersionResource{}
	if len(list) < 2 {
		panic("gvr input error, please input like format apps.v1.deployments or v1.pods")
	}
	// 区分
	if len(list) == 2 {
		ret.Version, ret.Resource = list[0], list[1]
	} else if len(list) > 2 {
		lastIndex := len(list) - 1
		ret.Version, ret.Resource = list[lastIndex-1], list[lastIndex]
		ret.Group = strings.Join(list[0:lastIndex-1], ".")
	}
	return ret
}

// isNameSpaceScope 是否 namespace 资源
func isNameSpaceScope(restMapper meta.RESTMapper, gvr schema.GroupVersionResource) bool {
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
