package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/myoperator/multiclusteroperator/pkg/server/service"
	"github.com/myoperator/multiclusteroperator/pkg/util"
	"io"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"strconv"
)

type ResourceController struct {
	ListService *service.ListService
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
	gvr := util.ParseIntoGvr(gvrParam, "/")
	if gvr.Empty() {
		c.JSON(400, gin.H{"error": "empty gvr input"})
		return
	}
	ll, _ := strconv.Atoi(limit)
	// 获取列表
	list, err := r.ListService.List(name, ns, cluster, labels, gvr, ll)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, list)
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
	gvr := util.ParseIntoGvr(gvrParam, "/")
	if gvr.Empty() {

	}
	ll, _ := strconv.Atoi(limit)
	// 获取列表
	list, err := r.ListService.ListWrapWithCluster(name, ns, cluster, labels, gvr, ll)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, list)
}

func (r *ResourceController) Join(c *gin.Context) {

	cluster := c.DefaultQuery("cluster", "")

	// 检查请求方法是否为 POST
	if c.Request.Method != http.MethodPost {
		http.Error(c.Writer, "只支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	// 解析上传的文件
	file, _, err := c.Request.FormFile("kubeconfig")
	if err != nil {
		http.Error(c.Writer, "无法解析上传的文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 创建临时文件用于保存上传的文件内容
	tempFile, err := ioutil.TempFile("", "kubeconfig-")
	if err != nil {
		http.Error(c.Writer, "无法创建临时文件", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// 将上传的文件内容写入临时文件
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(c.Writer, "无法保存上传的文件", http.StatusInternalServerError)
		return
	}

	// 使用 BuildConfigFromFlags 从上传的 kubeconfig 文件中获取配置
	config, err := clientcmd.BuildConfigFromFlags("", tempFile.Name())
	if err != nil {
		http.Error(c.Writer, fmt.Sprintf("无法获取 Kubernetes 配置: %v", err), http.StatusInternalServerError)
		return
	}

	// 获取列表
	err = r.ListService.Join(cluster, true, config)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, gin.H{"res": "join cluster successful"})
	return
}
