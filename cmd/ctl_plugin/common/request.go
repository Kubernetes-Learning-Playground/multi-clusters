package common

import (
	"bytes"
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"mime/multipart"
	"net/http"
	"os"
)

var (
	HttpClient     *Http
	ServerPort     int
	ServerIp       string
	KubeConfigPath string
)

func init() {
	HttpClient = &Http{
		Client: http.DefaultClient,
	}
	Cluster = ""
	Name = ""
}

var (
	Cluster string
	Name    string
	File    string
)

type Http struct {
	Client *http.Client
}

// createFormFile
// 将文件内容写入一个 multipart.Writer 中，并返回一个允许读取写入的文件内容的 io.Reader 对象。
// 发送 HTTP 请求时，我们需要将文件内容作为请求体的一部分进行发送。
func createFormFile(fieldName string, file *os.File) (io.Reader, *multipart.Writer) {
	bodyReader, bodyWriter := io.Pipe()
	writer := multipart.NewWriter(bodyWriter)

	go func() {
		defer bodyWriter.Close()

		part, err := writer.CreateFormFile(fieldName, file.Name())
		if err != nil {
			bodyWriter.CloseWithError(err)
			return
		}

		_, err = io.Copy(part, file)
		if err != nil {
			bodyWriter.CloseWithError(err)
			return
		}

		writer.Close()
	}()

	return bodyReader, writer
}

func (c *Http) DoUploadFile(url string, queryParams map[string]string, headerParams map[string]string, filePath string) ([]byte, error) {
	klog.V(4).Info("post url: ", url, " queryParams: ", queryParams, "header: ", headerParams)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body, writer := createFormFile("kubeconfig", file)

	// 创建 HTTP POST 请求
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		klog.Errorf("failed to create POST request: %v\n", err)
		return nil, err
	}
	params := req.URL.Query()
	for k, v := range queryParams {
		params.Add(k, v)
	}
	contentType := writer.FormDataContentType()
	req.Header.Set("Content-Type", contentType)
	req.URL.RawQuery = params.Encode()

	// 遍历头部信息，并添加到请求头中
	for key, value := range headerParams {
		req.Header.Set(key, value)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		klog.Errorf("failed to send POST request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP response error: %s\n, error: %v\n", resp.Status, err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("failed to read HTTP response body: %v\n", err)
		return nil, err
	}

	return responseBody, nil
}

func (c *Http) DoRequest(method string, url string, queryParams map[string]string, headerParams map[string]string, body []byte) ([]byte, error) {
	klog.V(4).Info(method, " url: ", url, " queryParams: ", queryParams)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		klog.Errorf("failed to create %s request: %v\n", method, err)
		return nil, err
	}
	params := req.URL.Query()
	for k, v := range queryParams {
		params.Add(k, v)
	}

	// 遍历头部信息，并添加到请求头中
	for key, value := range headerParams {
		req.Header.Set(key, value)
	}

	req.URL.RawQuery = params.Encode()

	resp, err := c.Client.Do(req)
	if err != nil {
		klog.Errorf("failed to send %s request: %v\n", method, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP response error: %s\n", resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("failed to read HTTP response body: %v\n", err)
		return nil, err
	}

	return responseBody, nil
}
