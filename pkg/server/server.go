package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/myoperator/multiclusteroperator/pkg/server/middleware"
	"github.com/myoperator/multiclusteroperator/pkg/server/service"
	"github.com/myoperator/multiclusteroperator/pkg/store"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

var (
	RR *ResourceController
)

type Server struct {
	factory store.Factory
	httpSrv *http.Server
}

func NewServer(addr int, tls *tls.Config) *Server {
	s := &http.Server{
		Addr: fmt.Sprintf(":%v", addr),
	}
	klog.Infof("http server port: %v\n", addr)
	if tls != nil {
		s.TLSConfig = tls
	}

	return &Server{
		httpSrv: s,
	}
}

func (s *Server) InjectStoreFactory(factory store.Factory) {
	s.factory = factory
}

func (s *Server) Start(db *gorm.DB) error {
	// route
	s.httpSrv.Handler = s.router(db)

	if s.httpSrv.TLSConfig == nil {
		return s.httpSrv.ListenAndServe()
	}
	return s.httpSrv.ListenAndServeTLS("", "")
}

func (s *Server) Stop() {
	if s.httpSrv != nil {
		s.httpSrv.Shutdown(context.TODO())
	}
}

func (s *Server) router(db *gorm.DB) http.Handler {
	r := gin.New()
	r.Use(middleware.LogMiddleware(), gin.Recovery())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Not Found"})
	})

	RR = &ResourceController{
		ListService: &service.ListService{
			DB: db,
		},
	}

	// pprof
	{
		r.GET("/debug/pprof/", gin.WrapF(pprof.Index))
		r.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
		r.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
		r.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
		r.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
	}

	v1 := r.Group("/v1")
	{
		v1.GET("/list", RR.List)
		v1.GET("/list_with_cluster", RR.ListWrapWithCluster)
	}

	return r
}
