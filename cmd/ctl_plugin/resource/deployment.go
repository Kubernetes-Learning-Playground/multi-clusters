package resource

import (
	"github.com/goccy/go-json"
	"github.com/olekukonko/tablewriter"
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"log"
	"os"
	"strconv"
)

var DeploymentCmd = &cobra.Command{}

func DeploymentCommand() *cobra.Command {

	DeploymentCmd = &cobra.Command{
		Use:          "deployments [flags]",
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

			err = ListDeployments(cluster, name, ns)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return DeploymentCmd

}

func ListDeployments(cluster, name, namespace string) error {

	m := map[string]string{}
	m["limit"] = "0"
	m["gvr"] = "apps.v1.deployments"
	if cluster != "" {
		m["cluster"] = cluster
	}

	if name != "" {
		m["name"] = name
	}

	if namespace != "" {
		m["namespace"] = namespace
	}

	rr := make([]*appsv1.Deployment, 0)
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
	content := []string{"集群名称", "Deployment", "Namespace", "TOTAL", "Available", "Ready"}

	//if common.ShowLabels {
	//	content = append(content, "标签")
	//}
	//if common.ShowAnnotations {
	//	content = append(content, "Annotations")
	//}

	table.SetHeader(content)

	for _, deployment := range rr {
		deploymentRow := []string{cluster, deployment.Name, deployment.Namespace, strconv.Itoa(int(deployment.Status.Replicas)), strconv.Itoa(int(deployment.Status.AvailableReplicas)), strconv.Itoa(int(deployment.Status.ReadyReplicas))}
		//podRow := []string{pod.Name, pod.Namespace, pod.Status.PodIP, string(pod.Status.Phase)}

		//if common.ShowLabels {
		//	podRow = append(podRow, common.LabelsMapToString(pod.Labels))
		//}
		//if common.ShowAnnotations {
		//	podRow = append(podRow, common.AnnotationsMapToString(pod.Annotations))
		//}

		table.Append(deploymentRow)
	}
	// 去掉表格线
	table = TableSet(table)

	table.Render()

	return nil

}
