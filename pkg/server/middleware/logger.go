package middleware

import (
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	"time"
)

// LogMiddleware 日志中间件
func LogMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()
		endTime := time.Now()
		// 响应时间
		execTime := endTime.Sub(startTime)
		requestMethod := ctx.Request.Method
		requestURI := ctx.Request.RequestURI
		statusCode := ctx.Writer.Status()
		requestIP := ctx.ClientIP()
		// 日志格式
		klog.Infof("| status=%2d | duration=%v | ip=%s | method=%s | url=%s |",
			statusCode,
			execTime,
			requestIP,
			requestMethod,
			requestURI,
		)
	}
}
