package main

import (
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	"github.com/practice/multi_resource/cmd/ctl_plugin/resource"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"log"
)

type CmdMetaData struct {
	Use     string
	Short   string
	Example string
}

var cmdMetaData *CmdMetaData

func init() {
	// FIXME: 要改
	cmdMetaData = &CmdMetaData{
		Use:     "kubectl list [flags]",
		Short:   "kubectl list resource",
		Example: "kubectl list [flags]",
	}

}

func main() {

	// 主命令
	mainCmd := &cobra.Command{
		Use:          cmdMetaData.Use,
		Short:        cmdMetaData.Short,
		Example:      cmdMetaData.Example,
		SilenceUsage: true,
	}

	// 各资源List命令
	podCmd := resource.PodCommand()
	depCmd := resource.DeploymentCommand()
	cmCmd := resource.ConfigmapCommand()

	// 注册
	MergeFlags(mainCmd, podCmd, depCmd, cmCmd)

	podCmd.Flags().StringVar(&common.Cluster, "clusterName", "", "")
	podCmd.Flags().StringVar(&common.Name, "name", "", "")

	depCmd.Flags().StringVar(&common.Cluster, "clusterName", "", "")
	depCmd.Flags().StringVar(&common.Name, "name", "", "")

	cmCmd.Flags().StringVar(&common.Cluster, "clusterName", "", "")
	cmCmd.Flags().StringVar(&common.Name, "name", "", "")

	// 主command需要加入子command
	mainCmd.AddCommand(podCmd, depCmd, cmCmd)

	err := mainCmd.Execute() // 主命令执行

	if err != nil {
		log.Fatalln(err)
	}

}

var cfgFlags *genericclioptions.ConfigFlags

func MergeFlags(cmds ...*cobra.Command) {
	cfgFlags = genericclioptions.NewConfigFlags(true)
	for _, cmd := range cmds {
		cfgFlags.AddFlags(cmd.Flags())
	}
}
