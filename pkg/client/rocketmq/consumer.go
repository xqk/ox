package rocketmq

import (
	"context"
	"ox/pkg/olog"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"ox/pkg/defers"
	"ox/pkg/istats"
)

type PushConsumer struct {
	rocketmq.PushConsumer
	name string
	ConsumerConfig

	subscribers  map[string]func(context.Context, ...*primitive.MessageExt) (consumer.ConsumeResult, error)
	interceptors []primitive.Interceptor
	fInfo        FlowInfo
}

func (conf *ConsumerConfig) Build() *PushConsumer {
	name := conf.Name
	if _, ok := _consumers.Load(name); ok {
		olog.Panic("duplicated load", olog.String("name", name))
	}

	olog.Debug("rocketmq's config: ", olog.String("name", name), olog.Any("conf", conf))

	pc := &PushConsumer{
		name:           name,
		ConsumerConfig: *conf,
		subscribers:    make(map[string]func(context.Context, ...*primitive.MessageExt) (consumer.ConsumeResult, error)),
		interceptors:   []primitive.Interceptor{},
		fInfo: FlowInfo{
			FlowInfoBase: istats.NewFlowInfoBase(conf.Shadow.Mode),
			Name:         name,
			Addr:         conf.Addr,
			Topic:        conf.Topic,
			Group:        conf.Group,
			GroupType:    "consumer",
		},
	}
	pc.interceptors = append(pc.interceptors, pushConsumerDefaultInterceptor(pc), pushConsumerMDInterceptor(pc), pushConsumerShadowInterceptor(pc, conf.Shadow))

	_consumers.Store(name, pc)
	return pc
}

func (cc *PushConsumer) Close() error {
	err := cc.Shutdown()
	if err != nil {
		olog.Warn("consumer close fail", olog.Any("error", err.Error()))
		return err
	}
	return nil
}

func (cc *PushConsumer) WithInterceptor(fs ...primitive.Interceptor) *PushConsumer {
	cc.interceptors = append(cc.interceptors, fs...)
	return cc
}

func (cc *PushConsumer) Subscribe(topic string, f func(context.Context, *primitive.MessageExt) error) *PushConsumer {
	if _, ok := cc.subscribers[topic]; ok {
		olog.Panic("duplicated subscribe", olog.String("topic", topic))
	}
	fn := func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			err := f(ctx, msg)
			if err != nil {
				olog.Error("consumer message", olog.String("err", err.Error()), olog.String("field", cc.name), olog.Any("ext", msg))
				return consumer.ConsumeRetryLater, err
			}
		}

		return consumer.ConsumeSuccess, nil
	}
	cc.subscribers[topic] = fn
	return cc
}

func (cc *PushConsumer) Start() error {
	// 初始化 PushConsumer
	client, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(cc.Group),
		consumer.WithNameServer(cc.Addr),
		consumer.WithMaxReconsumeTimes(cc.Reconsume),
		consumer.WithInterceptor(cc.interceptors...),
		consumer.WithCredentials(primitive.Credentials{
			AccessKey: cc.AccessKey,
			SecretKey: cc.SecretKey,
		}),
	)
	cc.PushConsumer = client

	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "",
	}
	if cc.ConsumerConfig.SubExpression != "*" {
		selector.Expression = cc.ConsumerConfig.SubExpression
	}

	for topic, fn := range cc.subscribers {
		if err := cc.PushConsumer.Subscribe(topic, selector, fn); err != nil {
			return err
		}
	}

	if err != nil || client == nil {
		olog.Panic("create consumer",
			olog.FieldName(cc.name),
			olog.FieldExtMessage(cc.ConsumerConfig),
			olog.Any("error", err),
		)
	}

	if cc.Enable {
		if err := client.Start(); err != nil {
			olog.Panic("start consumer",
				olog.FieldName(cc.name),
				olog.FieldExtMessage(cc.ConsumerConfig),
				olog.Any("error", err),
			)
			return err
		}
		// 在应用退出的时候，保证注销
		defers.Register(cc.Close)
	}

	return nil
}
