package multi_cluster_controller

import (
	"context"
	"fmt"
	"github.com/myoperator/multiclusteroperator/pkg/apis/multicluster/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ClusterHandler struct {
	// operator 控制器 client
	client.Client
	// 事件发送器
	EventRecorder record.EventRecorder
}

func NewClusterHandler(client client.Client, eventRecorder record.EventRecorder) *ClusterHandler {
	return &ClusterHandler{Client: client, EventRecorder: eventRecorder}
}

func (ch *ClusterHandler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// 获取 MultiCluster 资源对象
	rr := &v1alpha1.MultiCluster{}
	err := ch.Get(ctx, req.NamespacedName, rr)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if rr.Status.Status == "Healthy" {
		return reconcile.Result{}, nil
	}

	if !rr.DeletionTimestamp.IsZero() {
		klog.Infof("successful delete cluster %v\n", rr.Name)
		return reconcile.Result{}, nil
	}

	// 修改资源状态为 Healthy
	rr.Status.Status = "Healthy"

	err = ch.Client.Status().Update(ctx, rr)
	if err != nil {
		return reconcile.Result{}, err
	}

	ch.EventRecorder.Eventf(rr, corev1.EventTypeNormal, "Create", fmt.Sprintf("create %s cluster success", rr.Spec.Name))

	klog.Infof("successful reconcile...")
	return reconcile.Result{}, nil
}
