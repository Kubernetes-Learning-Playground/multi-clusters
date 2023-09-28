package helpers

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"testing"
)

func TestK8sApply(t *testing.T) {
	// 创建一个空的 Pod 对象
	pod := make(map[string]interface{})
	// 设置 Pod 的 apiVersion 和 kind
	pod["apiVersion"] = "v1"
	pod["kind"] = "Pod"
	// 设置 Pod 的元数据
	metaData := make(map[string]interface{})
	metaData["name"] = "my-pod"
	metaData["namespace"] = "default"
	pod["metadata"] = metaData

	// 设置 Pod 的规格
	spec := make(map[string]interface{})
	containers := []map[string]interface{}{
		{
			"name":  "my-container",
			"image": "nginx",
		},
	}
	spec["containers"] = containers
	pod["spec"] = spec
	// 将 Pod 对象转换为 JSON 字符串

	podJSON, err := json.Marshal(pod)
	if err != nil {
		fmt.Printf("Failed to convert Pod to JSON: %v\n", err)
		return
	}
	fmt.Println(string(podJSON))
	//_, err = K8sApply(podJSON, K8sRestConfig(), RestMapper())
	//if err != nil {
	//	log.Fatal(err)
	//}
	// 创建一个 Pod 对象
	podd := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "my-container",
					Image: "nginx",
				},
			},
		},
	}

	cc := InitClient()

	_, err = cc.CoreV1().Pods("default").Create(context.Background(), podd, metav1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}
	err = K8sDelete(podJSON, K8sRestConfig(), RestMapper())
	if err != nil {
		log.Fatal(err)
	}
}

// K8sRestConfig 初始化RestConfig配置
func K8sRestConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/zhenyu.jiang/.kube/config")
	if err != nil {
		log.Fatal(err)
	}
	return config
}

// RestMapper 获取所有资源对象 group-resource
// 初始化时先在内存保存，不需要重复从k8s中取
func RestMapper() meta.RESTMapper {
	gr, err := restmapper.GetAPIGroupResources(InitClient().Discovery())
	if err != nil {
		log.Fatal(err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)
	return mapper
}

// InitClient 初始化clientSet
func InitClient() *kubernetes.Clientset {
	c, err := kubernetes.NewForConfig(K8sRestConfig())
	if err != nil {
		log.Fatal(err)
	}
	return c
}
