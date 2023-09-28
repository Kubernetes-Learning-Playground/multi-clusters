package workqueue

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
)

type EventType string

var (
	AddEvent    EventType = "add"
	UpdateEvent EventType = "update"
	DeleteEvent EventType = "delete"
)

type QueueResource struct {
	Object    runtime.Object
	EventType EventType
}

// Queue 接口对象
type Queue interface {
	// Push 将监听到的资源放入queue中
	Push(resources *QueueResource)
	// Pop 拿出队列
	Pop() (*QueueResource, error)
	// ReQueue 重新放入队列，次数可配置
	ReQueue(*QueueResource) error
	// Finish 完成入列操作
	Finish(*QueueResource)
	// Close 关闭所有informer
	Close()
	// SetReMaxReQueueTime 设置最大重新入列次数
	SetReMaxReQueueTime(int)
}

// wq 使用限速队列实现queue接口
type wq struct {
	workqueue.RateLimitingInterface
	MaxReQueueTime int
}

var _ Queue = &wq{}

func NewWorkQueue(maxReQueueTime int) *wq {
	return &wq{
		workqueue.NewRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter()),
		maxReQueueTime,
	}
}

// Push 放入队列
func (c *wq) Push(obj *QueueResource) {
	c.AddRateLimited(obj)
}

// Pop 取出队列
func (c *wq) Pop() (*QueueResource, error) {
	obj, quit := c.Get()

	if quit {
		return &QueueResource{}, errors.New("Controller has been stopped.")
	}
	return obj.(*QueueResource), nil
}

// Finish 结束要干两件事，忘记+done
func (c *wq) Finish(obj *QueueResource) {
	c.Forget(obj)
	c.Done(obj)
}

// ReQueue 重新放入
func (c *wq) ReQueue(obj *QueueResource) error {
	if c.NumRequeues(obj) < c.MaxReQueueTime {
		// 这里会重新放入对列
		c.AddRateLimited(obj)
		return nil
	}
	// 如果次数大于最大重试次数，直接丢弃
	c.Forget(obj)
	c.Done(obj)
	return errors.New("This object has been requeued for many times, but still fails.")
}

func (c *wq) Close() {
	c.ShutDown()
}

func (c *wq) SetReMaxReQueueTime(maxReQueueTime int) {
	c.MaxReQueueTime = maxReQueueTime
}
