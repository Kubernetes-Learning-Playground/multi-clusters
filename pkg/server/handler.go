package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/myoperator/multiclusteroperator/pkg/server/service"
	"github.com/myoperator/multiclusteroperator/pkg/util"
	"io"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"strconv"
)

type ResourceController struct {
	ListService *service.ListService
	JoinService *service.JoinService
}

// List 查询接口
func (r *ResourceController) List(c *gin.Context) {
	// 接口传入参数
	gvrParam := c.Query("gvr")
	name := c.DefaultQuery("name", "")
	ns := c.DefaultQuery("namespace", "")
	limit := c.DefaultQuery("limit", "10")
	cluster := c.DefaultQuery("cluster", "")
	var labels map[string]string // 默认是nil
	if labelsQuery, ok := c.GetQueryMap("labels"); ok {
		labels = labelsQuery
	}

	// 解析 gvr
	gvr, _ := util.ParseIntoGvr(gvrParam, "/")
	if gvr.Empty() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not found gvr"})
		return
	}
	ll, _ := strconv.Atoi(limit)
	// 获取列表
	list, err := r.ListService.List(name, ns, cluster, labels, gvr, ll)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"list error": err})
		return
	}
	c.JSON(http.StatusOK, list)
	return
}

// ListWrapWithCluster 包裹 clusterName 返回，目前默认给命令行工具使用
func (r *ResourceController) ListWrapWithCluster(c *gin.Context) {
	// 接口传入参数
	gvrParam := c.Query("gvr")
	name := c.DefaultQuery("name", "")
	ns := c.DefaultQuery("namespace", "")
	limit := c.DefaultQuery("limit", "10")
	cluster := c.DefaultQuery("cluster", "")
	var labels map[string]string // 默认是nil
	if labelsQuery, ok := c.GetQueryMap("labels"); ok {
		labels = labelsQuery
	}

	// 解析 gvr
	gvr, _ := util.ParseIntoGvr(gvrParam, "/")
	if gvr.Empty() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not found gvr"})
		return
	}
	ll, _ := strconv.Atoi(limit)
	// 获取列表
	list, err := r.ListService.ListWrapWithCluster(name, ns, cluster, labels, gvr, ll)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"list wrap with cluster error": err})
		return
	}
	c.JSON(http.StatusOK, list)

}

func (r *ResourceController) Join(c *gin.Context) {

	cluster := c.DefaultQuery("cluster", "")
	insecure := c.DefaultQuery("insecure", "true")

	// 检查请求方法是否为 POST
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "only support POST request"})
		return
	}

	// 解析上传的文件
	file, _, err := c.Request.FormFile("kubeconfig")
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("parse file from request error: %s", err)})
		return
	}
	defer file.Close()

	// 创建临时文件用于保存上传的文件内容
	tempFile, err := os.CreateTemp("", "kubeconfig-")
	if err != nil {
		klog.Errorf("error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("create temp file error: %s", err)})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 将上传的文件内容写入临时文件
	_, err = io.Copy(tempFile, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("io copy error: %s", err)})
		return
	}

	// 使用 BuildConfigFromFlags 从上传的 kubeconfig 文件中获取配置
	config, err := clientcmd.BuildConfigFromFlags("", tempFile.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("get k8s config error: %s", err)})
		return
	}
	s, err := strconv.ParseBool(insecure)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	// 获取列表
	err = r.JoinService.Join(cluster, s, config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"join cluster error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"res": "join cluster successful"})
	return
}

func (r *ResourceController) UnJoin(c *gin.Context) {
	cluster := c.DefaultQuery("cluster", "")

	// 检查请求方法是否为 DELETE
	if c.Request.Method != http.MethodDelete {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "only support DELETE request"})
		return
	}

	// 获取列表
	err := r.JoinService.UnJoin(cluster)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"unjoin cluster error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"res": "unjoin cluster successful"})
	return
}
