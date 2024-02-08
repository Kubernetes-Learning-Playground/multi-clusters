package join

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	JoinCmd   *cobra.Command
	UnJoinCmd *cobra.Command
)

func init() {

	JoinCmd = &cobra.Command{
		Use:     "join <clusterName> [flags]",
		Short:   "join <clusterName> --file=xxx, input kubeconfig file path",
		Example: "join <clusterName> --file xxxxx",
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please specify a resource pods, deployments, configmaps\n")
			}
			clusterName := args[0]

			file, err := c.Flags().GetString("file")
			if err != nil {
				return err
			}

			err = joinClusterByFile(file, clusterName)
			if err != nil {
				return err
			}

			return nil
		},
	}

	UnJoinCmd = &cobra.Command{
		Use:     "unjoin <clusterName> [flags]",
		Short:   "unjoin <clusterName>, input clusterName",
		Example: "unjoin <clusterName>",
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("Please specify a resource pods, deployments, configmaps\n")
			}
			clusterName := args[0]
			err := unJoinClusterByName(clusterName)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
