package resource

import (
	"github.com/goccy/go-json"
	"github.com/olekukonko/tablewriter"
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"log"
	"os"
)

var PodCmd = &cobra.Command{}

func PodCommand() *cobra.Command {

	PodCmd = &cobra.Command{
		Use:          "pods [flags]",
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

			err = ListPods(cluster, name, ns)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return PodCmd

}

func ListPods(cluster, name, namespace string) error {

	m := map[string]string{}
	m["limit"] = "0"
	m["gvr"] = "v1.pods"
	if cluster != "" {
		m["cluster"] = cluster
	}

	if name != "" {
		m["name"] = name
	}

	if namespace != "" {
		m["namespace"] = namespace
	}

	rr := make([]*v1.Pod, 0)
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
	content := []string{"集群名称", "POD名称", "Namespace", "NODE", "POD IP", "状态", "容器名", "容器静像"}

	//if common.ShowLabels {
	//	content = append(content, "标签")
	//}
	//if common.ShowAnnotations {
	//	content = append(content, "Annotations")
	//}

	table.SetHeader(content)

	for _, pod := range rr {
		podRow := []string{cluster, pod.Name, pod.Namespace, pod.Spec.NodeName, pod.Status.PodIP, string(pod.Status.Phase), pod.Spec.Containers[0].Name, pod.Spec.Containers[0].Image}

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

func TableSet(table *tablewriter.Table) *tablewriter.Table {
	// 去掉表格线
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)

	return table
}
