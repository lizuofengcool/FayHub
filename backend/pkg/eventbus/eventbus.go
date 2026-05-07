package eventbus

import (
	"context"
	"sync"
	"time"

	"fayhub/pkg/logger"

	"go.uber.org/zap"
)

type Event struct {
	Name      string                 `json:"name"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	TenantID  int64                  `json:"tenant_id"`
}

type Handler func(ctx context.Context, event Event) error

type EventBus struct {
	handlers map[string][]Handler
	mu       sync.RWMutex
	queue    chan Event
	workers  int
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

var globalBus *EventBus

func Init(workers int) {
	if workers <= 0 {
		workers = 4
	}
	ctx, cancel := context.WithCancel(context.Background())
	globalBus = &EventBus{
		handlers: make(map[string][]Handler),
		queue:    make(chan Event, 1000),
		workers:  workers,
		ctx:      ctx,
		cancel:   cancel,
	}
	globalBus.start()
}

func GetBus() *EventBus {
	return globalBus
}

func (b *EventBus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

func (b *EventBus) Publish(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	select {
	case b.queue <- event:
	default:
		logger.Error(context.Background(), "事件队列已满，丢弃事件",
			zap.String("event", event.Name))
	}
}

func (b *EventBus) PublishAsync(eventName string, tenantID int64, payload map[string]interface{}) {
	b.Publish(Event{
		Name:     eventName,
		TenantID: tenantID,
		Payload:  payload,
	})
}

func (b *EventBus) start() {
	for i := 0; i < b.workers; i++ {
		b.wg.Add(1)
		go b.worker(i)
	}
}

func (b *EventBus) worker(id int) {
	defer b.wg.Done()
	for {
		select {
		case <-b.ctx.Done():
			return
		case event := <-b.queue:
			b.dispatch(event)
		}
	}
}

func (b *EventBus) dispatch(event Event) {
	b.mu.RLock()
	handlers, ok := b.handlers[event.Name]
	handlersCopy := make([]Handler, len(handlers))
	copy(handlersCopy, handlers)
	b.mu.RUnlock()

	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(b.ctx, 30*time.Second)
	defer cancel()

	for _, handler := range handlersCopy {
		if err := handler(ctx, event); err != nil {
			logger.Error(ctx, "事件处理器执行失败",
				zap.String("event", event.Name),
				zap.Error(err))
		}
	}
}

func (b *EventBus) Stop() {
	b.cancel()
	b.wg.Wait()
}

func Subscribe(eventName string, handler Handler) {
	if globalBus != nil {
		globalBus.Subscribe(eventName, handler)
	}
}

func PublishAsync(eventName string, tenantID int64, payload map[string]interface{}) {
	if globalBus != nil {
		globalBus.PublishAsync(eventName, tenantID, payload)
	}
}

const (
	EventUserCreated       = "user.created"
	EventUserDeleted       = "user.deleted"
	EventTenantCreated     = "tenant.created"
	EventTenantDeleted     = "tenant.deleted"
	EventPluginInstalled   = "plugin.installed"
	EventPluginUninstalled = "plugin.uninstalled"
	EventPaymentPaid       = "payment.paid"
	EventPaymentRefunded   = "payment.refunded"
	EventOrderExpired      = "order.expired"
	EventFileUploaded      = "file.uploaded"
	EventFileDeleted       = "file.deleted"
	EventLoginSuccess      = "auth.login.success"
	EventLoginFailed       = "auth.login.failed"
)

func RegisterBuiltinHandlers() {
	Subscribe(EventPaymentPaid, func(ctx context.Context, event Event) error {
		orderNo, _ := event.Payload["order_no"].(string)
		logger.Info(ctx, "支付成功事件",
			zap.String("event", EventPaymentPaid),
			zap.String("order_no", orderNo),
			zap.Int64("tenant_id", event.TenantID))
		return nil
	})

	Subscribe(EventOrderExpired, func(ctx context.Context, event Event) error {
		orderNo, _ := event.Payload["order_no"].(string)
		logger.Info(ctx, "订单过期事件",
			zap.String("event", EventOrderExpired),
			zap.String("order_no", orderNo))
		return nil
	})

	Subscribe(EventLoginFailed, func(ctx context.Context, event Event) error {
		username, _ := event.Payload["username"].(string)
		logger.Warn(ctx, "登录失败事件",
			zap.String("event", EventLoginFailed),
			zap.String("username", username))
		return nil
	})
}

type ScheduledTask struct {
	Name     string
	Interval time.Duration
	Fn       func(ctx context.Context) error
}

type Scheduler struct {
	tasks  []*ScheduledTask
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var globalScheduler *Scheduler

func InitScheduler() {
	ctx, cancel := context.WithCancel(context.Background())
	globalScheduler = &Scheduler{
		tasks:  make([]*ScheduledTask, 0),
		ctx:    ctx,
		cancel: cancel,
	}
}

func GetScheduler() *Scheduler {
	return globalScheduler
}

func (s *Scheduler) Register(task *ScheduledTask) {
	s.tasks = append(s.tasks, task)
}

func (s *Scheduler) Start() {
	for _, task := range s.tasks {
		s.wg.Add(1)
		go s.runTask(task)
	}
}

func (s *Scheduler) runTask(task *ScheduledTask) {
	defer s.wg.Done()
	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(s.ctx, task.Interval/2)
			if err := task.Fn(ctx); err != nil {
				logger.Error(ctx, "定时任务执行失败",
					zap.String("task", task.Name),
					zap.Error(err))
			}
			cancel()
		}
	}
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}

func RegisterTask(name string, interval time.Duration, fn func(ctx context.Context) error) {
	if globalScheduler != nil {
		globalScheduler.Register(&ScheduledTask{
			Name:     name,
			Interval: interval,
			Fn:       fn,
		})
	}
}

func StartScheduler() {
	if globalScheduler != nil {
		globalScheduler.Start()
	}
}

func StopAll() {
	if globalBus != nil {
		globalBus.Stop()
	}
	if globalScheduler != nil {
		globalScheduler.Stop()
	}
}

func InitAll() {
	Init(4)
	RegisterBuiltinHandlers()
	InitScheduler()
	StartScheduler()
}
