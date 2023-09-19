package resource

import (
	"github.com/goccy/go-json"
	"github.com/olekukonko/tablewriter"
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"log"
	"os"
	"strconv"
)

var ConfigmapCmd = &cobra.Command{}

func ConfigmapCommand() *cobra.Command {

	ConfigmapCmd = &cobra.Command{
		Use:          "configmaps [flags]",
		Short:        "list resource",
		Example:      "kubectl resource [flags]",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			cluster, err := c.Flags().GetString("clusterName")
			if err != nil {
				return err
			}

			ns, err := c.Flags().GetString("namespace")
			if err != nil {
				return err
			}

			name, err := c.Flags().GetString("name")
			if err != nil {
				return err
			}

			err = ListConfigmaps(cluster, name, ns)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return ConfigmapCmd

}

func ListConfigmaps(cluster, name, namespace string) error {

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

	rr := make([]*v1.ConfigMap, 0)
	r, err := common.HttpClient.DoGet("http://localhost:8888/v1/list", m)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(r, &rr)
	if err != nil {
		log.Fatal(err)
	}

	// 表格化呈现
	table := tablewriter.NewWriter(os.Stdout)
	content := []string{"集群名称", "Configmap", "Namespace", "DATA"}

	//if common.ShowLabels {
	//	content = append(content, "标签")
	//}
	//if common.ShowAnnotations {
	//	content = append(content, "Annotations")
	//}

	table.SetHeader(content)

	for _, cm := range rr {

		podRow := []string{cluster, cm.Name, cm.Namespace, strconv.Itoa(len(cm.Data))}
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
