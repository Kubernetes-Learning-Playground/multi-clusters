package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/practice/multi_resource/pkg/httpserver/service"
	"strconv"
)

type ResourceController struct {
	ListService *service.ListService
}

// List 对外暴露接口
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
	gvr := parseIntoGvr(gvrParam)
	if gvr.Empty() {
		panic("error gvr")
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
