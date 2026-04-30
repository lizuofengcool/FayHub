package messagequeue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fayhub/pkg/logger"
	"fayhub/pkg/redisclient"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Message struct {
	ID      string                 `json:"id"`
	Topic   string                 `json:"topic"`
	Payload map[string]interface{} `json:"payload"`
}

type ConsumerHandler func(ctx context.Context, msg Message) error

type MessageQueue struct {
	consumers map[string][]ConsumerHandler
	mu        sync.RWMutex
	groupName string
}

var globalMQ *MessageQueue

func Init() {
	globalMQ = &MessageQueue{
		consumers: make(map[string][]ConsumerHandler),
		groupName: "fayhub-mq",
	}
}

func GetMQ() *MessageQueue {
	return globalMQ
}

func (mq *MessageQueue) Publish(ctx context.Context, topic string, payload map[string]interface{}) error {
	client := redisclient.GetRawClient()
	if client == nil {
		return mq.publishFallback(topic, payload)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	msgID, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: mq.streamKey(topic),
		Values: map[string]interface{}{
			"topic":   topic,
			"payload": string(data),
		},
		ID:     "*",
		MaxLen: 10000,
		Approx: true,
	}).Result()
	if err != nil {
		return fmt.Errorf("发布消息到Redis Stream失败: %w", err)
	}

	logger.Info(ctx, "消息已发布到Redis Stream",
		zap.String("topic", topic),
		zap.String("msg_id", msgID))

	_ = msgID
	return nil
}

func (mq *MessageQueue) publishFallback(topic string, payload map[string]interface{}) error {
	msg := Message{
		ID:      fmt.Sprintf("local-%d", time.Now().UnixNano()),
		Topic:   topic,
		Payload: payload,
	}

	mq.mu.RLock()
	handlers, ok := mq.consumers[topic]
	handlersCopy := make([]ConsumerHandler, len(handlers))
	copy(handlersCopy, handlers)
	mq.mu.RUnlock()

	if !ok {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, handler := range handlersCopy {
		if err := handler(ctx, msg); err != nil {
			logger.Error(ctx, "本地消息处理失败",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}

	return nil
}

func (mq *MessageQueue) Subscribe(topic string, handler ConsumerHandler) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.consumers[topic] = append(mq.consumers[topic], handler)
}

func (mq *MessageQueue) StartConsumer(topic string) {
	client := redisclient.GetRawClient()
	if client == nil {
		return
	}

	streamKey := mq.streamKey(topic)

	mq.ensureGroup(context.Background(), client, streamKey)

	go mq.consumeLoop(client, streamKey, topic)
}

func (mq *MessageQueue) ensureGroup(ctx context.Context, client *redis.Client, streamKey string) {
	err := client.XGroupCreateMkStream(ctx, streamKey, mq.groupName, "0").Err()
	if err != nil {
		if err.Error() == "BUSYGROUP Consumer Group name already exists" {
			return
		}
		logger.Error(ctx, "创建Consumer Group失败",
			zap.String("stream", streamKey),
			zap.Error(err))
	}
}

func (mq *MessageQueue) consumeLoop(client *redis.Client, streamKey string, topic string) {
	consumerName := fmt.Sprintf("consumer-%d", time.Now().UnixNano())

	for {
		ctx := context.Background()

		results, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    mq.groupName,
			Consumer: consumerName,
			Streams:  []string{streamKey, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				continue
			}
			logger.Error(ctx, "从Redis Stream读取消息失败",
				zap.String("stream", streamKey),
				zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		for _, result := range results {
			for _, message := range result.Messages {
				mq.handleMessage(ctx, topic, message)
				client.XAck(ctx, streamKey, mq.groupName, message.ID)
			}
		}
	}
}

func (mq *MessageQueue) handleMessage(ctx context.Context, topic string, rawMsg redis.XMessage) {
	payloadStr, _ := rawMsg.Values["payload"].(string)
	if payloadStr == "" {
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		logger.Error(ctx, "反序列化消息失败",
			zap.String("msg_id", rawMsg.ID),
			zap.Error(err))
		return
	}

	msg := Message{
		ID:      rawMsg.ID,
		Topic:   topic,
		Payload: payload,
	}

	mq.mu.RLock()
	handlers, ok := mq.consumers[topic]
	handlersCopy := make([]ConsumerHandler, len(handlers))
	copy(handlersCopy, handlers)
	mq.mu.RUnlock()

	if !ok {
		return
	}

	for _, handler := range handlersCopy {
		if err := handler(ctx, msg); err != nil {
			logger.Error(ctx, "消息处理失败",
				zap.String("topic", topic),
				zap.String("msg_id", rawMsg.ID),
				zap.Error(err))
		}
	}
}

func (mq *MessageQueue) streamKey(topic string) string {
	return fmt.Sprintf("fayhub:mq:%s", topic)
}

func Publish(ctx context.Context, topic string, payload map[string]interface{}) error {
	if globalMQ != nil {
		return globalMQ.Publish(ctx, topic, payload)
	}
	return fmt.Errorf("消息队列未初始化")
}

func Subscribe(topic string, handler ConsumerHandler) {
	if globalMQ != nil {
		globalMQ.Subscribe(topic, handler)
	}
}

func StartConsumer(topic string) {
	if globalMQ != nil {
		globalMQ.StartConsumer(topic)
	}
}

func InitAndStart(topics []string) {
	Init()
	for _, topic := range topics {
		StartConsumer(topic)
	}
}
