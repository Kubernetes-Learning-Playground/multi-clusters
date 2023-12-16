package list

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/olekukonko/tablewriter"
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	"github.com/practice/multi_resource/pkg/store/model"
	"log"
	"os"
)

func Clusters(name string) error {
	m := map[string]string{}

	if name != "" {
		m["name"] = name
	}

	rr := make([]*model.Cluster, 0)
	url := fmt.Sprintf("http://localhost:%v/v1/list_cluster", common.ServerPort)
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
	content := []string{"集群名称", "状态", "是否为主集群"}

	table.SetHeader(content)

	for _, cm := range rr {

		podRow := []string{cm.Name, cm.Status, cm.IsMaster}

		table.Append(podRow)
	}
	// 去掉表格线
	table = TableSet(table)

	table.Render()

	return nil

}
