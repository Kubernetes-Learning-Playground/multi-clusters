package list

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	ListCmd *cobra.Command
)

func init() {

	ListCmd = &cobra.Command{
		Use:     "list resource [flags]",
		Short:   "list resource [flags]",
		Example: "list resources",
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please specify a resource pods, deployments, configmaps\n")
			}
			resource := args[0]

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

			switch resource {
			case "pods":
				err = Pods(cluster, name, ns)
			case "deployments":
				err = Deployments(cluster, name, ns)
			case "configmaps":
				err = Configmaps(cluster, name, ns)
			case "clusters":
				err = Clusters(name)
			default:
				return fmt.Errorf("Unsupport resource: %s\n", resource)
			}
			return err
		},
	}
}
