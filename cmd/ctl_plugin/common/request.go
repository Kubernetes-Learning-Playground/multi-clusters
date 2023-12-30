package common

import (
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"net/http"
)

var (
	HttpClient *Http
	ServerPort int
	ServerIp   string
)

func init() {
	HttpClient = &Http{
		Client: http.DefaultClient,
	}
	Cluster = ""
	Name = ""
}

var (
	Cluster       string
	Name          string
	CustomizePort string
)

type Http struct {
	Client *http.Client
}

func (c *Http) DoGet(url string, queryParams map[string]string) ([]byte, error) {
	// FIXME: 需要传入klog日志参数
	klog.V(4).Info("get url: ", url, " queryParams: ", queryParams)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	params := req.URL.Query()
	for k, v := range queryParams {
		params.Add(k, v)
	}
	req.URL.RawQuery = params.Encode()
	resp, err := c.Client.Do(req)
	if err != nil {
		klog.Errorf("client send err: %v\n", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http response error: %v\n", err)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		// 处理读取响应主体时的错误
		return nil, fmt.Errorf("http response readall error: %v\n", err)
	}

	return body, nil
}
