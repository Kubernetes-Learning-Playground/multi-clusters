package kubectl_client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/practice/multi_resource/pkg/kubectl_client/convert"
	"io"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util"
	"strings"
)

var Scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(scheme.AddToScheme(Scheme))
}

const DefaultDecoderBufferSize = 1024

// KubectlClient
type KubectlClient struct {
	DynamicClient   dynamic.Interface
	DiscoveryClient discovery.DiscoveryInterface
	Mapper          meta.RESTMapper
}

func NewKubectlManagerOrDie(config *rest.Config) *KubectlClient {

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	mapper, err := ToRESTMapper(discoveryClient)
	if err != nil {
		panic(err.Error())
	}

	return &KubectlClient{
		DynamicClient:   dynamicClient,
		DiscoveryClient: discoveryClient,
		Mapper:          mapper,
	}
}

// ToRESTMapper 获取 RESTMapper 用于 GVR <--> GVK 转换
func ToRESTMapper(discoveryClient discovery.DiscoveryInterface) (meta.RESTMapper, error) {
	gr, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)
	return mapper, nil
}

// Apply 模拟 kubectl apply 功能
// 1. 先转为 unstructList 对象
// 2. 遍历 apply 对象
func (o *KubectlClient) Apply(ctx context.Context, data []byte) error {

	// 解析为 unstruct 对象
	unstructObjList, err := decode(data)
	if err != nil {
		return err
	}

	for _, unstructObj := range unstructObjList {
		klog.Infof("Apply object: %#v", unstructObj)
		if _, err := o.applyUnstructured(ctx, unstructObj); err != nil {
			return err
		}
		klog.Infof("%s/%s applyed", strings.ToLower(unstructObj.GetKind()), unstructObj.GetName())
	}
	return nil
}

// Delete 模拟 kubectl delete 功能
// 1. 先转为 unstructList 对象
// 2. 遍历 delete 对象
func (o *KubectlClient) Delete(ctx context.Context, data []byte, isNotFoundErrIgnore bool) error {

	// 解析为 unstruct 对象
	unstructObjList, err := decode(data)
	if err != nil {
		return err
	}

	for _, unstructObj := range unstructObjList {
		klog.Infof("Delete object: %#v", unstructObj)
		if err := o.deleteUnstructured(ctx, unstructObj, isNotFoundErrIgnore); err != nil {
			return err
		}
		klog.Infof("%s/%s deleted", strings.ToLower(unstructObj.GetKind()), unstructObj.GetName())
	}
	return nil
}

// ApplyByResource 接受传入 interface{} 对象
func (o *KubectlClient) ApplyByResource(ctx context.Context, resource interface{}) error {
	data, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	return o.Apply(ctx, data)
}

// DeleteByResource 接受传入 interface{} 对象
func (o *KubectlClient) DeleteByResource(ctx context.Context, resource interface{}, isNotFoundErrIgnore bool) error {
	data, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	return o.Delete(ctx, data, isNotFoundErrIgnore)
}

// decode 解析传入 data: []byte --> []unstructured.Unstructured
// 可接受多对象传入，与 yaml 配置相同，需要使用 "---" 隔开
func decode(data []byte) ([]unstructured.Unstructured, error) {
	var lastErr error
	var unstructList []unstructured.Unstructured

	// 用于记录传入多少个对象
	i := 0

	// k8s 库中用于解析 yaml json 的解码器
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(data), DefaultDecoderBufferSize)

	for {
		// 内部解析需要使用 RawExtension 对象
		var reqObj runtime.RawExtension
		if err := decoder.Decode(&reqObj); err != nil {
			lastErr = err
			break
		}
		klog.Infof("The section:[%d] raw content: %s", i, string(reqObj.Raw))
		if len(reqObj.Raw) == 0 {
			continue
		}

		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(reqObj.Raw, nil, nil)
		if err != nil {
			lastErr = errors.Wrapf(err, "serialize the section:[%d] content error", i)
			klog.Info(lastErr)
			break
		}
		klog.Infof("The section:[%d] GroupVersionKind: %#v  object: %#v", i, gvk, obj)

		// 将 runtime.Object 对象 转换为 unstructured.Unstructured 对象
		unstruct, err := convert.ObjectToUnstructured(obj)
		if err != nil {
			lastErr = errors.Wrapf(err, "serialize the section:[%d] content error", i)
			break
		}
		unstructList = append(unstructList, unstruct)
		i++
	}

	if lastErr != io.EOF {
		return unstructList, errors.Wrapf(lastErr, "parsing the section:[%d] content error", i)
	}

	klog.Infof("object quantities:[%d]", i)

	return unstructList, nil
}

func (o *KubectlClient) applyUnstructured(ctx context.Context, unstructuredObj unstructured.Unstructured) (*unstructured.Unstructured, error) {

	if len(unstructuredObj.GetName()) == 0 {
		metadata, _ := meta.Accessor(unstructuredObj)
		generateName := metadata.GetGenerateName()
		if len(generateName) > 0 {
			return nil, fmt.Errorf("from %s: cannot use generate name with apply", generateName)
		}
	}

	b, err := unstructuredObj.MarshalJSON()
	if err != nil {
		return nil, err
	}

	obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(b, nil, nil)
	if err != nil {
		return nil, err
	}

	mapping, err := o.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	klog.Infof("mapping: %v", mapping.Scope.Name())

	var client dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if unstructuredObj.GetNamespace() == "" {
			unstructuredObj.SetNamespace("default")
		}
		client = o.DynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
	} else {
		client = o.DynamicClient.Resource(mapping.Resource)
	}

	// 需要对比 "last-applied-configuration" annotation 字段用
	modified, err := util.GetModifiedConfiguration(obj, true, unstructured.UnstructuredJSONScheme)
	if err != nil {
		return nil, fmt.Errorf("retrieving modified configuration from:\n%s\nfor:%v", unstructuredObj.GetName(), err)
	}

	// 先获取，如果没有直接创建
	currentUnstr, err := client.Get(ctx, unstructuredObj.GetName(), metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("retrieving current configuration of:\n%s\nfrom server for:%v", unstructuredObj.GetName(), err)
		}

		klog.Infof("The resource %s creating", unstructuredObj.GetName())
		// 创建 "last-applied-configuration" annotation
		if err := util.CreateApplyAnnotation(&unstructuredObj, unstructured.UnstructuredJSONScheme); err != nil {
			return nil, fmt.Errorf("creating %s error: %v", unstructuredObj.GetName(), err)
		}
		// create 操作
		return client.Create(ctx, &unstructuredObj, metav1.CreateOptions{})
	}

	klog.Infof("The resource already exists, so apply %s ", unstructuredObj.GetName())
	metadata, _ := meta.Accessor(currentUnstr)
	annotationMap := metadata.GetAnnotations()
	if _, ok := annotationMap[corev1.LastAppliedConfigAnnotation]; !ok {
		klog.Warningf("[%s] apply should be used on resource created by either kubectl create --save-config or apply", metadata.GetName())
	}

	// patch byte
	patchBytes, patchType, err := preparePatch(currentUnstr, modified, unstructuredObj.GetName(), *gvk)
	if err != nil {
		return nil, err
	}
	// patch 操作
	return client.Patch(ctx, unstructuredObj.GetName(), patchType, patchBytes, metav1.PatchOptions{})
}

func preparePatch(currentUnstr *unstructured.Unstructured, modified []byte, name string, gvk schema.GroupVersionKind) ([]byte, types.PatchType, error) {
	current, err := currentUnstr.MarshalJSON()
	if err != nil {
		return nil, "", fmt.Errorf("serializing current configuration from: %v, %v", currentUnstr, err)
	}

	original, err := util.GetOriginalConfiguration(currentUnstr)
	if err != nil {
		return nil, "", fmt.Errorf("retrieving original configuration from: %s, %v", name, err)
	}

	var patchType types.PatchType
	var patch []byte

	versionedObject, err := Scheme.New(gvk)
	switch {
	case runtime.IsNotRegisteredError(err):
		patchType = types.MergePatchType
		preconditions := []mergepatch.PreconditionFunc{
			mergepatch.RequireKeyUnchanged("apiVersion"),
			mergepatch.RequireKeyUnchanged("kind"),
			mergepatch.RequireKeyUnchanged("name"),
		}
		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
		if err != nil {
			if mergepatch.IsPreconditionFailed(err) {
				return nil, "", fmt.Errorf("At least one of apiVersion, kind and name was changed")
			}
			return nil, "", fmt.Errorf("unable to apply patch, %v", err)
		}
	case err == nil:
		patchType = types.StrategicMergePatchType
		lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(versionedObject)
		if err != nil {
			return nil, "", err
		}
		patch, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true)
		if err != nil {
			return nil, "", err
		}
	case err != nil:
		return nil, "", fmt.Errorf("getting instance of versioned object %v for: %v", gvk, err)
	}

	return patch, patchType, nil
}

func (o *KubectlClient) deleteUnstructured(ctx context.Context, unstructuredObj unstructured.Unstructured, isNotFoundErrIgnore bool) error {

	if len(unstructuredObj.GetName()) == 0 {
		metadata, _ := meta.Accessor(unstructuredObj)
		generateName := metadata.GetGenerateName()
		if len(generateName) > 0 {
			return fmt.Errorf("from %s: cannot use generate name with delete", generateName)
		}
	}

	b, err := unstructuredObj.MarshalJSON()
	if err != nil {
		return err
	}

	_, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(b, nil, nil)
	if err != nil {
		return err
	}

	mapping, err := o.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}
	klog.Infof("mapping: %v", mapping.Scope.Name())

	var client dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if unstructuredObj.GetNamespace() == "" {
			unstructuredObj.SetNamespace("default")
		}
		client = o.DynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
	} else {
		client = o.DynamicClient.Resource(mapping.Resource)
	}

	// delete 操作
	err = client.Delete(ctx, unstructuredObj.GetName(), metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		if isNotFoundErrIgnore {
			klog.Warningf("resource: %v/%v not found", unstructuredObj.GetKind(), unstructuredObj.GetName())
			return nil
		}
		return fmt.Errorf("resource: %v/%v not found", unstructuredObj.GetKind(), unstructuredObj.GetName())
	}

	return err
}
