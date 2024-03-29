package list

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/myoperator/multiclusteroperator/cmd/ctl_plugin/common"
	"github.com/olekukonko/tablewriter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"net/http"
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
	url := fmt.Sprintf("http://%v:%v/v1/list_with_cluster", common.ServerIp, common.ServerPort)
	r, err := common.HttpClient.DoRequest(http.MethodGet, url, m, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(r, &rr)
	if err != nil {
		log.Fatal(err)
	}

	// 表格化呈现
	table := tablewriter.NewWriter(os.Stdout)
	content := []string{"Cluster", "Name", "Namespace"}

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
