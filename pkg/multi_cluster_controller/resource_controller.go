package multi_cluster_controller

import (
	"context"
	"fmt"
	"github.com/practice/multi_resource/pkg/apis/multiclusterresource/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

func (mc *MultiClusterHandler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// 获取 Resource
	rr := &v1alpha1.MultiClusterResource{}
	err := mc.Get(ctx, req.NamespacedName, rr)
	mc.Logger.Info("get multi-cluster-resource: ", rr.GetName()+"/"+rr.GetNamespace())
	if err != nil {
		if errors.IsNotFound(err) {
			mc.Logger.Info("not found multi-cluster-resource: ", rr.GetName()+"/"+rr.GetNamespace())
			return reconcile.Result{}, nil
		}
		mc.Logger.Error(err, "get multi-cluster-resource: ", rr.GetName()+"/"+rr.GetNamespace(), " failed")
		return reconcile.Result{}, err
	}

	// 删除状态，会等到 Finalizer 字段清空后才会真正删除
	// 1、删除所有集群资源
	// 2、清空 Finalizer，更新状态
	if !rr.DeletionTimestamp.IsZero() {
		err = mc.resourceDelete(rr)
		if err != nil {
			mc.Logger.Error(err, "delete multi-cluster-resource: ", rr.GetName()+"/"+rr.GetNamespace(), " failed")
			mc.EventRecorder.Event(rr, corev1.EventTypeWarning, "Delete", fmt.Sprintf("delete %s fail", rr.Name))
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 60}, err
		}

		err = mc.Client.Update(ctx, rr)
		if err != nil {
			mc.Logger.Error(err, "update multi-cluster-resource: ", rr.GetName()+"/"+rr.GetNamespace(), " failed")
			mc.EventRecorder.Event(rr, corev1.EventTypeWarning, "UpdateFailed", fmt.Sprintf("update %s fail", rr.Name))
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 60}, err
		}

		return reconcile.Result{}, nil
	}

	// 设置 crd 对象的 Finalizer 字段，并判断是否改变
	forDelete, finalizer, isChange := mc.setResourceFinalizer(rr)

	// 如果 Finalizer 字段改变，
	// 代表可能是需要进行特定集群的删除资源操作
	if isChange {
		err = mc.resourceDeleteBySlice(rr, forDelete)
		if err != nil {
			mc.Logger.Error(err, "delete slice multi-cluster-resource: ", rr.GetName()+"/"+rr.GetNamespace(), " failed")
			mc.EventRecorder.Event(rr, corev1.EventTypeWarning, "DeleteFailed", fmt.Sprintf("resourceDeleteBySlice %s fail", rr.Name))
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 60}, err
		}
		// 删除后覆盖
		rr.Finalizers = finalizer
		err = mc.Client.Update(ctx, rr)
		if err != nil {
			mc.Logger.Error(err, "update slice multi-cluster-resource: ", rr.GetName()+"/"+rr.GetNamespace(), " failed")
			mc.EventRecorder.Event(rr, corev1.EventTypeWarning, "UpdateFailed", fmt.Sprintf("update %s fail", rr.Name))
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 60}, err
		}
	}

	// apply 操作
	err = mc.resourceApply(rr)
	if err != nil {
		mc.Logger.Error(err, "apply slice multi-cluster-resource: ", rr.GetName()+"/"+rr.GetNamespace(), " failed")
		mc.EventRecorder.Event(rr, corev1.EventTypeWarning, "ApplyFailed", fmt.Sprintf("resourceApply %s fail", rr.Name))
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 60}, err
	}

	// patch 操作
	err = mc.resourcePatch(rr)
	if err != nil {
		mc.Logger.Error(err, "patch slice multi-cluster-resource: ", rr.GetName()+"/"+rr.GetNamespace(), " failed")
		mc.EventRecorder.Event(rr, corev1.EventTypeWarning, "PatchFailed", fmt.Sprintf("resourcePatch %s fail", rr.Name))
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 60}, err
	}

	mc.Logger.Info("reconcile multi-cluster-resource: ", rr.GetName()+"/"+rr.GetNamespace(), " success")

	return reconcile.Result{}, nil
}
