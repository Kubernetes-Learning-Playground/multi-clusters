package list

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/olekukonko/tablewriter"
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"os"
)

type WrapResource struct {
	Object      *unstructured.Unstructured `json:"Object"`
	ClusterName string                     `json:"clusterName"`
}

func Resources(cluster, name, namespace, gvr string) error {

	m := map[string]string{}
	m["limit"] = "0"
	m["gvr"] = gvr
	if cluster != "" {
		m["cluster"] = cluster
	}

	if name != "" {
		m["name"] = name
	}

	if namespace != "" {
		m["namespace"] = namespace
	}

	rr := make([]*WrapResource, 0)
	url := fmt.Sprintf("http://localhost:%v/v1/list_with_cluster", common.ServerPort)
	r, err := common.HttpClient.DoGet(url, m)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(r, &rr)
	if err != nil {
		log.Fatal(err)
	}

	// 表格化呈现
	table := tablewriter.NewWriter(os.Stdout)
	content := []string{"集群名称", "Name", "Namespace"}

	table.SetHeader(content)

	for _, pod := range rr {
		pod.Object.GetResourceVersion()
		podRow := []string{pod.ClusterName, pod.Object.GetName(), pod.Object.GetNamespace()}

		table.Append(podRow)
	}
	// 去掉表格线
	table = TableSet(table)

	table.Render()

	return nil

}
