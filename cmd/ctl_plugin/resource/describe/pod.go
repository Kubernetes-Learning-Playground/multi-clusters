package describe

import (
	"fmt"
	yy "github.com/ghodss/yaml"
	"github.com/goccy/go-json"
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	v1 "k8s.io/api/core/v1"
	"log"
)

func Pods(cluster, name, namespace string) error {

	m := map[string]string{}
	m["limit"] = "0"
	m["gvr"] = "v1.pods"
	if cluster != "" {
		m["cluster"] = cluster
	}

	if name != "" {
		m["name"] = name
	}

	if namespace != "" {
		m["namespace"] = namespace
	}

	rr := make([]*v1.Pod, 0)
	url := fmt.Sprintf("http://%v:%v/v1/list", common.ServerIp, common.ServerPort)
	r, err := common.HttpClient.DoGet(url, m)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(r, &rr)
	if err != nil {
		log.Fatal(err)
	}

	for _, pod := range rr {
		resByte, err := json.Marshal(pod)
		if err != nil {
			log.Fatal(err)
		}
		resByte, _ = yy.JSONToYAML(resByte)
		fmt.Printf(string(resByte))
		fmt.Println("---------------------------------")
	}

	return nil

}
