package caches

import (
	"context"
	"github.com/practice/multi_resource/pkg/store"
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
	store.Queue
}

// Start 启动工作队列
func (r *ResourceHandler) Start(ctx context.Context) {
	klog.Info("worker queue start...")
	go func() {
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

// HandleObject 处理 work queue 传入对象
func (r *ResourceHandler) handleObject(obj *store.QueueResource) error {
	klog.Infof("handler [%s] object from work queue\n", obj.EventType)
	res, err := store.NewResource(obj.Object, r.RestMapper)
	if err != nil {
		klog.Errorf("new resource [%s] object error: %s\n", obj.EventType, err)
		return err
	}

	// 区分传入的事件，并进行相应处理
	switch obj.EventType {
	case store.AddEvent:

		err = res.Add(r.DB)
		if err != nil {
			klog.Errorf("[%s] object error: %s\n", obj.EventType, err)
			return err
		}

		return nil
	case store.UpdateEvent:

		err = res.Update(r.DB)
		if err != nil {
			klog.Errorf("[%s] object error: %s\n", obj.EventType, err)
			return err
		}

		return nil
	case store.DeleteEvent:

		err = res.Delete(r.DB)
		if err != nil {
			klog.Errorf("[%s] object error: %s\n", obj.EventType, err)
			return err
		}
	}
	return nil
}

func NewResourceHandler(DB *gorm.DB, restMapper meta.RESTMapper) *ResourceHandler {
	return &ResourceHandler{
		DB:         DB,
		RestMapper: restMapper,
		Queue:      store.NewWorkQueue(5),
	}
}

func (r *ResourceHandler) OnAdd(obj interface{}, isInInitialList bool) {
	klog.Info("add resource...")
	if o, ok := obj.(runtime.Object); ok {
		rr := &store.QueueResource{Object: o, EventType: store.AddEvent}
		r.Push(rr)
	}

}

func (r *ResourceHandler) OnUpdate(oldObj, newObj interface{}) {
	klog.Info("update resource...")
	if o, ok := newObj.(runtime.Object); ok {
		rr := &store.QueueResource{Object: o, EventType: store.UpdateEvent}
		r.Push(rr)
	}
}

func (r *ResourceHandler) OnDelete(obj interface{}) {
	klog.Info("delete resource...")
	if o, ok := obj.(runtime.Object); ok {
		rr := &store.QueueResource{Object: o, EventType: store.DeleteEvent}
		r.Push(rr)
	}
}

var _ cache.ResourceEventHandler = &ResourceHandler{}
