package convert

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ObjectToUnstructuredList 转换 runtime.Object -> []unstructured.Unstructured
func ObjectToUnstructuredList(obj runtime.Object) ([]unstructured.Unstructured, error) {
	list := make([]unstructured.Unstructured, 0, 0)
	if meta.IsListType(obj) {
		if _, ok := obj.(*unstructured.UnstructuredList); !ok {
			return nil, fmt.Errorf("unable to convert runtime object to list")
		}

		for _, u := range obj.(*unstructured.UnstructuredList).Items {
			list = append(list, u)
		}
		return list, nil
	}

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	unstructuredObj := unstructured.Unstructured{Object: unstructuredMap}
	list = append(list, unstructuredObj)

	return list, nil
}

// ObjectToUnstructured 转换 runtime.Object -> unstructured.Unstructured
func ObjectToUnstructured(obj runtime.Object) (unstructured.Unstructured, error) {
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	return unstructured.Unstructured{Object: unstructuredMap}, nil
}
