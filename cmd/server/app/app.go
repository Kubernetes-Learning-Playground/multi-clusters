package app

import (
	"context"
	"github.com/myoperator/multiclusteroperator/cmd/server/app/options"
	"github.com/myoperator/multiclusteroperator/pkg/leaselock"
	"github.com/myoperator/multiclusteroperator/pkg/multi_cluster_controller"
	"github.com/myoperator/multiclusteroperator/pkg/util"
	"github.com/spf13/cobra"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
	"k8s.io/klog/v2"
	"os"
	"time"
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

			if err := leaderElectionRun(opts); err != nil {
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

var podName = os.Getenv("POD_NAME")

// leaderElectionRun 是否启动选举机制
func leaderElectionRun(opts *options.Options) error {

	switch opts.Server.LeaseLockMode {
	case true:
		klog.Info("lead election mode run...")
		var config *rest.Config
		if opts.Server.LeaseLockMode {
			config, _ = rest.InClusterConfig()
		}

		client := clientset.NewForConfigOrDie(config)

		lock := leaselock.GetNewLock(opts.Server.LeaseLockName, podName, opts.Server.LeaseLockNamespace, client)
		// 选主模式
		leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
			Lock:            lock,
			ReleaseOnCancel: true,
			LeaseDuration:   15 * time.Second, // 租约时长，follower用来判断集群锁是否过期
			RenewDeadline:   10 * time.Second, // leader更新锁的时长
			RetryPeriod:     2 * time.Second,  // 重试获取锁的间隔
			// 当发生不同选主事件时的回调方法
			Callbacks: leaderelection.LeaderCallbacks{
				// 成为leader时，需要执行的回调
				OnStartedLeading: func(c context.Context) {
					// 执行server逻辑
					klog.Info("leader election server running...")
					err := run(opts)
					if err != nil {
						return
					}
				},
				// 不是leader时，需要执行的回调
				OnStoppedLeading: func() {
					klog.Info("no longer a leader...")
					klog.Info("clean up server...")
				},
				// 当产生新leader时，执行的回调
				OnNewLeader: func(currentId string) {
					if currentId == podName {
						klog.Info("still the leader!")
						return
					}
					klog.Infof("new leader is %v", currentId)
				},
			},
		})

	case false:
		klog.Info("normal run...")
		err := run(opts)
		if err != nil {
			return err
		}
	}
	return nil
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
		return err
	}

	// 4. 启动多集群 handler
	mch.StartWorkQueueHandler(ctx)

	// 5. 启动 operator 管理器
	go func() {
		defer util.HandleCrash()
		klog.Info("operator manager start...")
		if err = mch.StartOperatorManager(); err != nil {
			klog.Fatal(err)
		}
	}()

	// 6. 启动 http server
	return server.Start(mysqlClient.GetDB())
}
