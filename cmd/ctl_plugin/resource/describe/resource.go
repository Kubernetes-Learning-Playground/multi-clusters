package describe

import (
	"fmt"
	yy "github.com/ghodss/yaml"
	"github.com/goccy/go-json"
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
)

// FIXME: describe 命令可以直接解析 gvk 就好，不用特别分资源

func Resource(cluster, name, namespace, gvr string) error {

	m := map[string]string{}
	m["limit"] = "0"
	m["gvr"] = gvr
	if cluster != "" {
		m["cluster"] = cluster
	}

	if name != "" {
		m["name"] = name
	}

	if namespace != "" {
		m["namespace"] = namespace
	}

	rr := make([]*unstructured.Unstructured, 0)
	url := fmt.Sprintf("http://%v:%v/v1/list", common.ServerIp, common.ServerPort)
	r, err := common.HttpClient.DoGet(url, m)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(r, &rr)
	if err != nil {
		log.Fatal(err)
	}

	for _, re := range rr {
		resByte, err := json.Marshal(re)
		if err != nil {
			log.Fatal(err)
		}
		resByte, _ = yy.JSONToYAML(resByte)
		fmt.Printf(string(resByte))
		fmt.Println("---------------------------------")
	}

	return nil

}
