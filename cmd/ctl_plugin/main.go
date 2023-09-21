package main

import (
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	"github.com/practice/multi_resource/cmd/ctl_plugin/resource/describe"
	"github.com/practice/multi_resource/cmd/ctl_plugin/resource/list"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"log"
)

type CmdMetaData struct {
	Use     string
	Short   string
	Example string
}

var (
	cmdMetaData *CmdMetaData
)

func init() {

	cmdMetaData = &CmdMetaData{
		Use:     "kubectl [flags]",
		Short:   "kubectl [flags]",
		Example: "kubectl [flags]",
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

	// 注册
	MergeFlags(list.ListCmd, describe.DescribeCmd)

	list.ListCmd.Flags().StringVar(&common.Cluster, "clusterName", "", "")
	list.ListCmd.Flags().StringVar(&common.Name, "name", "", "")

	describe.DescribeCmd.Flags().StringVar(&common.Cluster, "clusterName", "", "")
	describe.DescribeCmd.Flags().StringVar(&common.Name, "name", "", "")

	// 主command需要加入子command
	mainCmd.AddCommand(list.ListCmd, describe.DescribeCmd)

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
