package helpers

import (
	"bytes"
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/util"
	"log"
)

func setDefaultNamespaceIfScopedAndNoneSet(u *unstructured.Unstructured, helper *resource.Helper) {
	namespace := u.GetNamespace()
	if helper.NamespaceScoped && namespace == "" {
		namespace = "default"
		u.SetNamespace(namespace)
	}
}

func newRestClient(restConfig *rest.Config, gv schema.GroupVersion) (rest.Interface, error) {
	restConfig.ContentConfig = resource.UnstructuredPlusDefaultContentConfig()
	restConfig.GroupVersion = &gv
	if len(gv.Group) == 0 {
		restConfig.APIPath = "/api"
	} else {
		restConfig.APIPath = "/apis"
	}

	return rest.RESTClientFor(restConfig)
}

func K8sDelete(json []byte, restConfig *rest.Config, mapper meta.RESTMapper) error {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(json),
		len(json))
	for {
		var rawObj runtime.RawExtension
		err := decoder.Decode(&rawObj)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		// 得到gvk
		obj, gvk, err := syaml.NewDecodingSerializer(unstructured.
			UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			log.Fatal(err)
		}

		//把obj 变成map[string]interface{}
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return nil
		}
		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		restMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		restClient, err := newRestClient(restConfig, gvk.GroupVersion())

		helper := resource.NewHelper(restClient, restMapping)

		setDefaultNamespaceIfScopedAndNoneSet(unstructuredObj, helper)

		_, err = helper.Delete(unstructuredObj.GetNamespace(), unstructuredObj.GetName())
		if err != nil {
			log.Println(fmt.Sprintf("删除资源 %s/%s 失败:%s", unstructuredObj.GetNamespace(),
				unstructuredObj.GetName(), err.Error(),
			))
		}

	}
	return nil
}

func K8sApply(json []byte, restConfig *rest.Config, mapper meta.RESTMapper) ([]*resource.Info, error) {
	resList := []*resource.Info{}

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(json), len(json))

	for {
		var rawObj runtime.RawExtension
		err := decoder.Decode(&rawObj)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return resList, err
			}
		}
		// 得到gvk
		obj, gvk, err := syaml.NewDecodingSerializer(unstructured.
			UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return resList, err
		}
		//把obj 变成map[string]interface{}
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return resList, err
		}
		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		restMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return resList, err
		}

		restClient, err := newRestClient(restConfig, gvk.GroupVersion())

		helper := resource.NewHelper(restClient, restMapping)

		setDefaultNamespaceIfScopedAndNoneSet(unstructuredObj, helper)

		objInfo := &resource.Info{
			Client:          restClient,
			Mapping:         restMapping,
			Namespace:       unstructuredObj.GetNamespace(),
			Name:            unstructuredObj.GetName(),
			Object:          unstructuredObj,
			ResourceVersion: restMapping.Resource.Version,
		}

		patcher, err := NewPatcher(objInfo, helper)
		if err != nil {
			return resList, err
		}

		modified, err := util.GetModifiedConfiguration(objInfo.Object, true, unstructured.UnstructuredJSONScheme)
		if err != nil {
			return resList, err
		}

		if err := objInfo.Get(); err != nil {
			if !errors.IsNotFound(err) { //资源不存在
				return resList, err
			}

			if err := util.CreateApplyAnnotation(objInfo.Object, unstructured.UnstructuredJSONScheme); err != nil {
				return resList, err
			}

			// 直接创建
			obj, err := helper.Create(objInfo.Namespace, true, objInfo.Object)
			if err != nil {

				fmt.Println("有错")
				return resList, err
			}
			objInfo.Refresh(obj, true)
		}

		_, patchedObject, err := patcher.Patch(objInfo.Object, modified, objInfo.Namespace, objInfo.Name)
		if err != nil {
			return resList, err
		}

		objInfo.Refresh(patchedObject, true)

		// 把ObjectInfo 塞入列表
		resList = append(resList, objInfo)
	}
	return resList, nil
}
