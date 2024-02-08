package main

import (
	"github.com/myoperator/multiclusteroperator/cmd/ctl_plugin/common"
	"github.com/myoperator/multiclusteroperator/cmd/ctl_plugin/resource/describe"
	"github.com/myoperator/multiclusteroperator/cmd/ctl_plugin/resource/join"
	"github.com/myoperator/multiclusteroperator/cmd/ctl_plugin/resource/list"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/create"
	"k8s.io/kubectl/pkg/cmd/delete"
	"k8s.io/kubectl/pkg/cmd/edit"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"os"
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
		Short:   "kubectl-multicluster, Multi-cluster command line tool",
		Example: "kubectl-multicluster list core/v1/pods, or kubectl-multicluster describe core/v1/pods",
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
	common.KubeConfigPath = r.MasterClusterKubeConfigPath
	// 注册 list describe 命令
	MergeFlags(list.ListCmd, describe.DescribeCmd, join.JoinCmd, join.UnJoinCmd)
	// 只需要加入 --clusterName=xxx, --name=xxx, 其他适配 kubectl
	list.ListCmd.Flags().StringVar(&common.Cluster, "clusterName", "", "--clusterName=xxx")
	list.ListCmd.Flags().StringVar(&common.Name, "name", "", "--name=xxx")

	describe.DescribeCmd.Flags().StringVar(&common.Cluster, "clusterName", "", "--clusterName=xxx")
	describe.DescribeCmd.Flags().StringVar(&common.Name, "name", "", "--name=xxx")
	// join 命令需要 --file=xxx 上传文件
	join.JoinCmd.Flags().StringVar(&common.File, "file", "", "--file=xxx")

	// kubeconfig 配置
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	// 获取clientSet
	kubeConfigFlags.KubeConfig = &common.KubeConfigPath
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	// 输出地点
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	// 主 Cmd 需要加入子 Cmd
	mainCmd.AddCommand(list.ListCmd,
		describe.DescribeCmd,
		join.JoinCmd,
		join.UnJoinCmd,
		apply.NewCmdApply("kubectl", f, ioStreams),
		delete.NewCmdDelete(f, ioStreams),
		create.NewCmdCreate(f, ioStreams),
		edit.NewCmdEdit(f, ioStreams),
	)

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
