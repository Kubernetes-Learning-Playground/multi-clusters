package app

import (
	"context"
	"github.com/myoperator/multiclusteroperator/cmd/server/app/options"
	"github.com/myoperator/multiclusteroperator/pkg/config"
	"github.com/myoperator/multiclusteroperator/pkg/multi_cluster_controller"
	"github.com/myoperator/multiclusteroperator/pkg/util"
	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
	"k8s.io/klog/v2"
)

func NewServerCommand() *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Use: "multi-clusters-server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliflag.PrintFlags(cmd.Flags())

			if err := opts.Complete(); err != nil {
				klog.Errorf("unable to complete options, %+v", err)
				return err
			}

			if err := opts.Validate(); err != nil {
				klog.Errorf("unable to validate options, %+v", err)
				return err
			}

			if err := run(opts); err != nil {
				klog.Errorf("unable to run server, %+v", err)
				return err
			}

			return nil
		},
	}

	fs := cmd.Flags()
	namedFlagSets := opts.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cliflag.SetUsageAndHelpFunc(cmd, namedFlagSets, cols)

	return cmd
}

// run 启动 http server + operator manager
func run(opts *options.Options) error {
	// 1. 初始化 db 实例
	mysqlClient, err := opts.MySQL.NewClient()
	if err != nil {
		return err
	}

	// 2. 实例化 server
	server, err := opts.Server.NewServer()
	if err != nil {
		return err
	}

	// 3. 注入 db factory
	server.InjectStoreFactory(mysqlClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mch, err := multi_cluster_controller.NewMultiClusterHandlerFromConfig(opts.Server.ConfigPath, mysqlClient.GetDB())
	if err != nil {
		klog.Fatal(err)
	}

	// 4. 创建 .multi-cluster-operator config 文件
	config.CreateCtlFile(opts.Server, mch.MasterClusterKubeConfigPath)

	// 5. 启动多集群 handler
	mch.StartWorkQueueHandler(ctx)

	// 6. 启动 operator 管理器
	go func() {
		defer util.HandleCrash()
		klog.Info("operator manager start...")
		if err = mch.StartOperatorManager(); err != nil {
			klog.Fatal(err)
		}
	}()

	// 7. 启动 http server
	return server.Start(mysqlClient.GetDB())
}
