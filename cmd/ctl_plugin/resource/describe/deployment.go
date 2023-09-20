package describe

import (
	"fmt"
	yy "github.com/ghodss/yaml"
	"github.com/goccy/go-json"
	"github.com/practice/multi_resource/cmd/ctl_plugin/common"
	appsv1 "k8s.io/api/apps/v1"
	"log"
)

func Deployments(cluster, name, namespace string) error {

	m := map[string]string{}
	m["limit"] = "0"
	m["gvr"] = "apps.v1.deployments"
	if cluster != "" {
		m["cluster"] = cluster
	}

	if name != "" {
		m["name"] = name
	}

	if namespace != "" {
		m["namespace"] = namespace
	}

	rr := make([]*appsv1.Deployment, 0)
	r, err := common.HttpClient.DoGet("http://localhost:8888/v1/list", m)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(r, &rr)
	if err != nil {
		log.Fatal(err)
	}

	for _, deployment := range rr {
		resByte, err := json.Marshal(deployment)
		if err != nil {
			log.Fatal(err)
		}
		resByte, _ = yy.JSONToYAML(resByte)
		fmt.Printf(string(resByte))
		fmt.Println("---------------------------------")
	}

	return nil

}
