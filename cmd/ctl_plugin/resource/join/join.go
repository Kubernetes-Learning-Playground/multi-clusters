package join

import (
	"fmt"
	"github.com/myoperator/multiclusteroperator/cmd/ctl_plugin/common"
	"net/http"
)

// joinClusterByFile 根据传入文件调用 join 接口
func joinClusterByFile(filePath, clusterName string) error {
	m := map[string]string{}

	if clusterName != "" {
		m["cluster"] = clusterName
	}

	url := fmt.Sprintf("http://%v:%v/v1/join", common.ServerIp, common.ServerPort)
	b, err := common.HttpClient.DoUploadFile(url, m, nil, filePath)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// unJoinClusterByName 根据传入集群名调用 unjoin 接口
func unJoinClusterByName(clusterName string) error {
	m := map[string]string{}

	if clusterName != "" {
		m["cluster"] = clusterName
	}

	url := fmt.Sprintf("http://%v:%v/v1/unjoin", common.ServerIp, common.ServerPort)
	b, err := common.HttpClient.DoRequest(http.MethodDelete, url, m, nil, nil)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
