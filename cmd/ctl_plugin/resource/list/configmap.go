package list

import (
	"github.com/goccy/go-json"
	"github.com/olekukonko/tablewriter"
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	v1 "k8s.io/api/core/v1"
	"log"
	"os"
	"strconv"
)

type WrapConfigMap struct {
	Object      *v1.ConfigMap `json:"Object"`
	ClusterName string        `json:"clusterName"`
}

func Configmaps(cluster, name, namespace string) error {

	m := map[string]string{}
	m["limit"] = "0"
	m["gvr"] = "v1.configmaps"
	if cluster != "" {
		m["cluster"] = cluster
	}

	if name != "" {
		m["name"] = name
	}

	if namespace != "" {
		m["namespace"] = namespace
	}

	rr := make([]*WrapConfigMap, 0)
	r, err := common.HttpClient.DoGet("http://localhost:8888/v1/list_with_cluster", m)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(r, &rr)
	if err != nil {
		log.Fatal(err)
	}

	// 表格化呈现
	table := tablewriter.NewWriter(os.Stdout)
	content := []string{"集群名称", "Name", "Namespace", "DATA"}

	//if common.ShowLabels {
	//	content = append(content, "标签")
	//}
	//if common.ShowAnnotations {
	//	content = append(content, "Annotations")
	//}

	table.SetHeader(content)

	for _, cm := range rr {

		podRow := []string{cm.ClusterName, cm.Object.Name, cm.Object.Namespace, strconv.Itoa(len(cm.Object.Data))}
		//podRow := []string{pod.Name, pod.Namespace, pod.Status.PodIP, string(pod.Status.Phase)}

		//if common.ShowLabels {
		//	podRow = append(podRow, common.LabelsMapToString(pod.Labels))
		//}
		//if common.ShowAnnotations {
		//	podRow = append(podRow, common.AnnotationsMapToString(pod.Annotations))
		//}

		table.Append(podRow)
	}
	// 去掉表格线
	table = TableSet(table)

	table.Render()

	return nil

}
