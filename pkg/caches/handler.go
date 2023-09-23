package caches

import (
	"context"
	"github.com/practice/multi_resource/pkg/caches/workqueue"
	"github.com/practice/multi_resource/pkg/store"
	"github.com/practice/multi_resource/pkg/util"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type ResourceHandler struct {
	DB *gorm.DB
	// RestMapper 资源对象
	RestMapper meta.RESTMapper
	// Queue 工作队列接口
	workqueue.Queue
	// clusterName 集群名
	clusterName string
}

// Start 启动工作队列
func (r *ResourceHandler) Start(ctx context.Context) {
	klog.Info("worker queue start...")
	go func() {
		defer util.HandleCrash()
		for {
			select {
			case <-ctx.Done():
				klog.Info("exit work queue...")
				r.Close()
				return
			default:
			}

			// 不断由队列中获取元素处理
			obj, err := r.Pop()
			if err != nil {
				klog.Errorf("work queue pop error: %s\n", err)
				continue
			}

			// 如果自己的业务逻辑发生问题，可以重新放回队列。
			if err = r.handleObject(obj); err != nil {
				klog.Errorf("handle obj from work queue error: %s\n", err)
				// 重新入列
				_ = r.ReQueue(obj)
			} else {
				// 完成就结束
				r.Finish(obj)
			}
		}
	}()
}

// handleObject 处理 work queue 传入对象
func (r *ResourceHandler) handleObject(obj *workqueue.QueueResource) error {
	//klog.Infof("[%s] handler [%s] object from work queue\n", r.clusterName, obj.EventType)
	res, err := store.NewResource(obj.Object, r.RestMapper, r.clusterName)
	if err != nil {
		klog.Errorf("new resource [%s] object error: %s\n", obj.EventType, err)
		return err
	}

	// 区分传入的事件，并进行相应处理
	switch obj.EventType {
	case workqueue.AddEvent:
		err = res.Add(r.DB)
		if err != nil {
			klog.Errorf("[%s] [%s] object error: %s\n", r.clusterName, obj.EventType, err)
			return err
		}

	case workqueue.UpdateEvent:
		err = res.Update(r.DB)
		if err != nil {
			klog.Errorf("[%s] [%s] object error: %s\n", r.clusterName, obj.EventType, err)
			return err
		}

	case workqueue.DeleteEvent:
		err = res.Delete(r.DB)
		if err != nil {
			klog.Errorf("[%s] [%s] object error: %s\n", r.clusterName, obj.EventType, err)
			return err
		}
	}
	return nil
}

func NewResourceHandler(DB *gorm.DB, restMapper meta.RESTMapper, clusterName string) *ResourceHandler {
	return &ResourceHandler{
		DB:          DB,
		RestMapper:  restMapper,
		Queue:       workqueue.NewWorkQueue(5),
		clusterName: clusterName,
	}
}

func (r *ResourceHandler) OnAdd(obj interface{}, isInInitialList bool) {
	if o, ok := obj.(runtime.Object); ok {
		rr := &workqueue.QueueResource{Object: o, EventType: workqueue.AddEvent}
		r.Push(rr)
	}
}

func (r *ResourceHandler) OnUpdate(oldObj, newObj interface{}) {
	if o, ok := newObj.(runtime.Object); ok {
		rr := &workqueue.QueueResource{Object: o, EventType: workqueue.UpdateEvent}
		r.Push(rr)
	}
}

func (r *ResourceHandler) OnDelete(obj interface{}) {
	if o, ok := obj.(runtime.Object); ok {
		rr := &workqueue.QueueResource{Object: o, EventType: workqueue.DeleteEvent}
		r.Push(rr)
	}
}

var _ cache.ResourceEventHandler = &ResourceHandler{}
