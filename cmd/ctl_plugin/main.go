package main

import (
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	"github.com/practice/multi_resource/cmd/ctl_plugin/resource/describe"
	"github.com/practice/multi_resource/cmd/ctl_plugin/resource/list"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
	"strconv"
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
		Use:     "kubectl-multicluster [flags]",
		Short:   "kubectl-multicluster [flags]",
		Example: "kubectl-multicluster [flags]",
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

	// 从配置文件获取端口信息
	r := common.LoadConfigFile()
	common.ServerPort, _ = strconv.Atoi(r.ServerPort)
	common.ServerIp = r.ServerIP
	// 注册 list describe 命令
	MergeFlags(list.ListCmd, describe.DescribeCmd)
	// 只需要加入 --clusterName=xxx, --name=xxx, 其他适配 kubectl
	list.ListCmd.Flags().StringVar(&common.Cluster, "clusterName", "", "--clusterName=xxx")
	list.ListCmd.Flags().StringVar(&common.Name, "name", "", "--name=xxx")

	describe.DescribeCmd.Flags().StringVar(&common.Cluster, "clusterName", "", "--clusterName=xxx")
	describe.DescribeCmd.Flags().StringVar(&common.Name, "name", "", "--name=xxx")

	// 主command需要加入子command
	mainCmd.AddCommand(list.ListCmd, describe.DescribeCmd)

	err := mainCmd.Execute() // 主命令执行

	if err != nil {
		klog.Fatalln(err)
	}
}

var cfgFlags *genericclioptions.ConfigFlags

func MergeFlags(cmds ...*cobra.Command) {
	cfgFlags = genericclioptions.NewConfigFlags(true)
	for _, cmd := range cmds {
		cfgFlags.AddFlags(cmd.Flags())
	}
}
