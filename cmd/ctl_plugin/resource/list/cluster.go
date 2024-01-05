package list

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/myoperator/multiclusteroperator/cmd/ctl_plugin/common"
	"github.com/myoperator/multiclusteroperator/pkg/apis/multicluster/v1alpha1"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
)

type WrapCluster struct {
	Object      *v1alpha1.MultiCluster `json:"Object"`
	ClusterName string                 `json:"clusterName"`
}

func Clusters(name string) error {
	m := map[string]string{}

	if name != "" {
		m["name"] = name
	}

	m["limit"] = "0"
	m["gvr"] = "mulitcluster.practice.com/v1alpha1/multiclusters"

	m["namespace"] = "default"

	rr := make([]*WrapCluster, 0)
	url := fmt.Sprintf("http://%v:%v/v1/list_with_cluster", common.ServerIp, common.ServerPort)
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
	content := []string{"Name", "VERSION", "HOST", "PLATFORM", "ISMASTER", "STATUS"}

	table.SetHeader(content)

	for _, cm := range rr {

		podRow := []string{cm.Object.Name, cm.Object.Spec.Version, cm.Object.Spec.Host, cm.Object.Spec.Platform, cm.Object.Spec.IsMaster, cm.Object.Status.Status}

		table.Append(podRow)
	}
	// 去掉表格线
	table = TableSet(table)

	table.Render()

	return nil

}
