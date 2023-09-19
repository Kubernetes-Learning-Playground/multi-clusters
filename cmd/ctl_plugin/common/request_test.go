package common

import (
	"fmt"
	v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"testing"
)

func TestRequest(t *testing.T) {

	m := map[string]string{}
	m["gvr"] = "v1.configmaps"
	m["cluster"] = "cluster1"
	rr := make([]*v1.ConfigMap, 0)
	r, err := HttpClient.DoGet("http://localhost:8888/v1/list", m)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(r, &rr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rr)
}
