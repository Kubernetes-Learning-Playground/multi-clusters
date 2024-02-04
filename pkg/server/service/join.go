package service

import (
	"github.com/myoperator/multiclusteroperator/pkg/config"
	"github.com/myoperator/multiclusteroperator/pkg/multi_cluster_controller"
	"k8s.io/client-go/rest"
)

type JoinService struct {
	Mch *multi_cluster_controller.MultiClusterHandler
}

func (list *ListService) Join(clusterName string, insecure bool, restConfig *rest.Config) error {

	cluster := &config.Cluster{
		MetaData: config.MetaData{
			ClusterName: clusterName,
			IsMaster:    false,
			Insecure:    true,
			RestConfig:  restConfig,
		},
	}

	return multi_cluster_controller.AddMultiClusterHandler(cluster)
}
