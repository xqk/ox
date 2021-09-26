package rocketmq

import (
	"context"
	"ox/pkg/olog"
	"ox/pkg/util/odebug"
	"strings"
	"time"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"ox/pkg/imeta"
	"ox/pkg/istats"
	"ox/pkg/metric"
)

type FlowInfo struct {
	Name      string   `json:"name"`
	Addr      []string `json:"addr"`
	Topic     string   `json:"topic"`
	Group     string   `json:"group"`
	GroupType string   `json:"groupType"` // 类型， consumer 消费者， producer 生产者
	istats.FlowInfoBase
}

func consumeResultStr(result consumer.ConsumeResult) string {
	switch result {
	case consumer.ConsumeSuccess:
		return "success"
	case consumer.ConsumeRetryLater:
		return "retryLater"
	case consumer.Commit:
		return "commit"
	case consumer.Rollback:
		return "rollback"
	case consumer.SuspendCurrentQueueAMoment:
		return "suspendCurrentQueueAMoment"
	default:
		return "unknown"
	}
}

func pushConsumerDefaultInterceptor(pushConsumer *PushConsumer) primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		beg := time.Now()
		msgs := req.([]*primitive.MessageExt)
		err := next(ctx, msgs, reply)
		if reply == nil {
			return err
		}

		holder := reply.(*consumer.ConsumeResultHolder)
		odebug.PrintObject("consume", map[string]interface{}{
			"err":    err,
			"count":  len(msgs),
			"result": consumeResultStr(holder.ConsumeResult),
		})

		// 消息处理结果统计
		for _, msg := range msgs {
			host := msg.StoreHost
			topic := msg.Topic
			result := consumeResultStr(holder.ConsumeResult)
			if err != nil {
				olog.Error("push consumer",
					olog.String("topic", topic),
					olog.String("host", host),
					olog.String("result", result),
					olog.Any("err", err))

			} else {
				olog.Info("push consumer",
					olog.String("topic", topic),
					olog.String("host", host),
					olog.String("result", result),
				)
			}
			metric.ClientHandleCounter.Inc(metric.TypeRocketMQ, topic, "consume", host, result)
			metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.TypeRocketMQ, topic, "consume", host)
		}
		if pushConsumer.RwTimeout > time.Duration(0) {
			if time.Since(beg) > pushConsumer.RwTimeout {
				olog.Error("slow",
					olog.String("topic", pushConsumer.Topic),
					olog.String("result", consumeResultStr(holder.ConsumeResult)),
					olog.Any("cost", time.Since(beg).Seconds()),
				)
			}
		}

		return err
	}
}

func pushConsumerMDInterceptor(pushConsumer *PushConsumer) primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		msgs := req.([]*primitive.MessageExt)
		if len(msgs) > 0 {
			var meta = imeta.New(nil)
			for key, vals := range msgs[0].GetProperties() {
				if strings.HasPrefix(strings.ToLower(key), "x-dy") {
					meta.Set(key, strings.Split(vals, ",")...)
				}
			}
			ctx = imeta.WithContext(ctx, meta)
		}
		err := next(ctx, msgs, reply)
		return err
	}
}

func pushConsumerShadowInterceptor(pushConsumer *PushConsumer, config Shadow) primitive.Interceptor {
	isWitheTopic := func(topicName string) bool {
		for _, v := range config.WitheTopics {
			if topicName == v {
				return true
			}
		}
		return false
	}
	addr := strings.Join(pushConsumer.Addr, ",")
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		msg := req.([]*primitive.MessageExt)
		pushConsumer.fInfo.UpdateFlow()
		if len(msg) > 0 {
			realReq := msg[0]
			switch config.Mode {
			case "off":
			case "on":
				if md, ok := imeta.FromContext(ctx); ok && md.IsShadow() {
					pushConsumer.fInfo.UpdateShadowFlow()
					if !isWitheTopic(realReq.Topic) {
						olog.Info(
							"SHADOW_DROP_MSG",
							olog.FieldAddr(addr),
							olog.FieldMethod(realReq.Topic),
							olog.String("body", string(realReq.Body)),
							olog.FieldType("consumer"),
						)
						return nil
					}
				}
			case "watch":
				var wouldDrop bool
				if md, ok := imeta.FromContext(ctx); ok && md.IsShadow() && !isWitheTopic(realReq.Topic) {
					wouldDrop = true
				}
				olog.Info("SHADOW_WATCH_MSG",
					olog.FieldAddr(addr),
					olog.FieldMethod(realReq.Topic),
					olog.Any("wouldDrop", wouldDrop),
					olog.Any("WitheTopics", config.WitheTopics),
					olog.FieldType("consumer"),
				)
			}
		}

		err := next(ctx, req, reply)
		return err
	}
}

func produceResultStr(result primitive.SendStatus) string {
	switch result {
	case primitive.SendOK:
		return "sendOk"
	case primitive.SendFlushDiskTimeout:
		return "sendFlushDiskTimeout"
	case primitive.SendFlushSlaveTimeout:
		return "sendFlushSlaveTimeout"
	case primitive.SendSlaveNotAvailable:
		return "sendSlaveNotAvailable"
	default:
		return "unknown"
	}
}

func producerDefaultInterceptor(producer *Producer) primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		beg := time.Now()
		realReq := req.(*primitive.Message)
		realReply := reply.(*primitive.SendResult)
		err := next(ctx, realReq, realReply)
		if realReply == nil || realReply.MessageQueue == nil {
			return err
		}

		odebug.PrintObject("produce", map[string]interface{}{
			"err":     err,
			"message": realReq,
			"result":  realReply.String(),
		})

		// 消息处理结果统计
		topic := producer.Topic
		if err != nil {
			olog.Error("produce",
				olog.String("topic", topic),
				olog.String("queue", ""),
				olog.String("result", realReply.String()),
				olog.Any("err", err),
			)
			metric.ClientHandleCounter.Inc(metric.TypeRocketMQ, topic, "produce", "unknown", err.Error())
			metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.TypeRocketMQ, topic, "produce", "unknown")
		} else {
			olog.Info("produce",
				olog.String("topic", topic),
				olog.Any("queue", realReply.MessageQueue),
				olog.String("result", produceResultStr(realReply.Status)),
			)
			metric.ClientHandleCounter.Inc(metric.TypeRocketMQ, topic, "produce", realReply.MessageQueue.BrokerName, produceResultStr(realReply.Status))
			metric.ClientHandleHistogram.Observe(time.Since(beg).Seconds(), metric.TypeRocketMQ, topic, "produce", realReply.MessageQueue.BrokerName)
		}

		if producer.RwTimeout > time.Duration(0) {
			if time.Since(beg) > producer.RwTimeout {
				olog.Error("slow",
					olog.String("topic", topic),
					olog.String("result", realReply.String()),
					olog.Any("cost", time.Since(beg).Seconds()),
				)
			}
		}

		return err
	}
}

// 统一minerva metadata 传递
func producerMDInterceptor(producer *Producer) primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		if md, ok := imeta.FromContext(ctx); ok {
			realReq := req.(*primitive.Message)
			for k, v := range md {
				realReq.WithProperty(k, strings.Join(v, ","))
			}
		}
		err := next(ctx, req, reply)
		return err
	}
}

func producerShadowInterceptor(producer *Producer, config Shadow) primitive.Interceptor {
	isWitheTopic := func(topicName string) bool {
		for _, v := range config.WitheTopics {
			if topicName == v {
				return true
			}
		}
		return false
	}
	addr := strings.Join(producer.Addr, ",")
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		realReq := req.(*primitive.Message)
		producer.fInfo.UpdateFlow()
		switch config.Mode {
		case "off":
		case "on":
			if md, ok := imeta.FromContext(ctx); ok && md.IsShadow() {
				producer.fInfo.UpdateShadowFlow()
				if !isWitheTopic(realReq.Topic) {
					// 压测模式非白名单topic直接丢弃
					olog.Info(
						"SHADOW_DROP_MSG",
						olog.FieldAddr(addr),
						olog.FieldMethod(realReq.Topic),
						olog.String("body", string(realReq.Body)),
						olog.FieldType("producer"),
					)
					return nil
				}
			}
		case "watch":
			var wouldDrop bool
			if md, ok := imeta.FromContext(ctx); ok && md.IsShadow() && !isWitheTopic(realReq.Topic) {
				wouldDrop = true
			}
			olog.Info("SHADOW_WATCH_MSG",
				olog.FieldAddr(addr),
				olog.FieldMethod(realReq.Topic),
				olog.Any("wouldDrop", wouldDrop),
				olog.Any("WitheTopics", config.WitheTopics),
				olog.FieldType("producer"),
			)
		}
		err := next(ctx, req, reply)
		return err
	}
}
