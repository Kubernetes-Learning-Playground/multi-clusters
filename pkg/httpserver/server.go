package httpserver

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/practice/multi_resource/pkg/config"
	"github.com/practice/multi_resource/pkg/httpserver/service"
	"k8s.io/klog/v2"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

var (
	RR *ResourceController
)

func HttpServer(ctx context.Context, opt *config.Options, dp *config.Dependencies) error {

	if !opt.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// 心跳检测健康机制
	go func() {
		h := &healthz.Handler{
			Checks: map[string]healthz.Checker{
				"healthz": healthz.Ping,
			},
		}
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", opt.HealthPort), h); err != nil {
			klog.Fatalf("Failed to start healthz endpoint: %v", err)
		}
	}()

	RR = &ResourceController{
		ListService: &service.ListService{
			DB: dp.DB,
		},
	}

	router := gin.New()
	router.Use(gin.Recovery())

	errCh := make(chan error, 1)

	register(router)
	err := router.Run(fmt.Sprintf(":%v", opt.Port))
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		// 上下文被取消，关闭服务器并返回上下文错误
		return ctx.Err()
	case err := <-errCh:
		// 从错误通道获取错误信息，并将错误传递给上下文对象 ctx
		return fmt.Errorf("internal error: %w", err)
	}
}

func register(router *gin.Engine) {

	r := router.Group("/v1")
	{
		r.GET("/list_cluster", RR.ListCluster)
		r.GET("/list", RR.List)
		r.GET("/list_with_cluster", RR.ListWrapWithCluster)
	}
}
