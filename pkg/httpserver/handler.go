package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/practice/multi_resource/pkg/httpserver/service"
	"github.com/practice/multi_resource/pkg/util"
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
	gvr := util.ParseIntoGvr(gvrParam, ".")
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
	gvr := util.ParseIntoGvr(gvrParam, ".")
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

// ListCluster 查询接口
func (r *ResourceController) ListCluster(c *gin.Context) {
	// 接口传入参数
	name := c.DefaultQuery("name", "")
	limit := c.DefaultQuery("limit", "10")

	ll, _ := strconv.Atoi(limit)
	// 获取列表
	list, err := r.ListService.ListCluster(name, ll)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, list)
	return
}

