package rocketmq

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/xqk/ox/pkg/defers"
	"github.com/xqk/ox/pkg/istats"
	"github.com/xqk/ox/pkg/olog"
	"github.com/xqk/ox/pkg/util/odebug"
)

type Producer struct {
	rocketmq.Producer
	name string
	ProducerConfig
	interceptors []primitive.Interceptor
	fInfo        FlowInfo
}

func StdNewProducer(name string) *Producer {
	return StdProducerConfig("configName").Build()
}

func (conf *ProducerConfig) Build() *Producer {
	name := conf.Name
	if _, ok := _producers.Load(name); ok {
		olog.Panic("duplicated load", olog.String("name", name))
	}

	if odebug.IsDevelopmentMode() {
		odebug.PrettyJsonPrint("rocketmq's config: "+name, conf)
	}

	cc := &Producer{
		name:           name,
		ProducerConfig: *conf,
		interceptors:   []primitive.Interceptor{},
		fInfo: FlowInfo{
			FlowInfoBase: istats.NewFlowInfoBase(conf.Shadow.Mode),
			Name:         name,
			Addr:         conf.Addr,
			Topic:        conf.Topic,
			Group:        conf.Group,
			GroupType:    "producer",
		},
	}

	cc.interceptors = append(cc.interceptors, producerDefaultInterceptor(cc), producerMDInterceptor(cc), producerShadowInterceptor(cc, conf.Shadow))

	_producers.Store(name, cc)
	return cc
}

func (pc *Producer) Start() error {
	// 兼容配置
	client, err := rocketmq.NewProducer(
		producer.WithNameServer(pc.Addr),
		producer.WithRetry(pc.Retry),
		producer.WithInterceptor(pc.interceptors...),
		producer.WithInstanceName(pc.name),
		producer.WithCredentials(primitive.Credentials{
			AccessKey: pc.AccessKey,
			SecretKey: pc.SecretKey,
		}),
	)
	if err != nil || client == nil {
		olog.Panic("create producer",
			olog.FieldName(pc.name),
			olog.FieldExtMessage(pc.ProducerConfig),
			olog.Any("error", err),
		)
	}

	if err := client.Start(); err != nil {
		olog.Panic("start producer",
			olog.FieldName(pc.name),
			olog.FieldExtMessage(pc.ProducerConfig),
			olog.Any("error", err),
		)
	}

	pc.Producer = client
	// 在应用退出的时候，保证注销
	defers.Register(pc.Close)
	return nil
}

func (pc *Producer) WithInterceptor(fs ...primitive.Interceptor) *Producer {
	pc.interceptors = append(pc.interceptors, fs...)
	return pc
}

func (pc *Producer) Close() error {
	err := pc.Shutdown()
	if err != nil {
		olog.Warn("consumer close fail", olog.Any("error", err.Error()))
		return err
	}
	_producers.Delete(pc.name)
	return nil
}

// Send rocketmq发送消息
func (pc *Producer) Send(msg []byte) error {
	m := primitive.NewMessage(pc.Topic, msg)
	_, err := pc.SendSync(context.Background(), m)
	if err != nil {
		olog.Error("send message error", olog.Any("msg", msg))
		return err
	}
	return nil
}

// SendWithContext 发送消息
func (pc *Producer) SendWithContext(ctx context.Context, msg []byte) error {
	m := primitive.NewMessage(pc.Topic, msg)
	_, err := pc.SendSync(ctx, m)
	if err != nil {
		olog.Error("send message error", olog.Any("msg", msg))
		return err
	}
	return nil
}

// SendWithTag rocket mq 发送消息,可以自定义选择 tag
func (pc *Producer) SendWithTag(msg []byte, tag string) error {
	m := primitive.NewMessage(pc.Topic, msg)
	if tag != "" {
		m.WithTag(tag)
	}

	_, err := pc.SendSync(context.Background(), m)
	if err != nil {
		olog.Error("send message error", olog.Any("msg", msg))
		return err
	}
	return nil
}

// SendWithResult rocket mq 发送消息,可以自定义选择 tag 及返回结果
func (pc *Producer) SendWithResult(msg []byte, tag string) (*primitive.SendResult, error) {
	m := primitive.NewMessage(pc.Topic, msg)
	if tag != "" {
		m.WithTag(tag)
	}

	res, err := pc.SendSync(context.Background(), m)
	if err != nil {
		olog.Error("send message error", olog.Any("msg", msg))
		return res, err
	}
	return res, nil
}

// SendMsg... 自定义消息格式
func (pc *Producer) SendMsg(msg *primitive.Message) (*primitive.SendResult, error) {
	res, err := pc.SendSync(context.Background(), msg)
	if err != nil {
		olog.Error("send message error", olog.Any("msg", msg))
		return res, err
	}
	return res, nil
}
