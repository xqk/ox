package client

import "github.com/xqk/ox/pkg/client/rocketmq"

type Config struct {
	EndpointProducer rocketmq.ProducerConfig `conf:"producer" gen:"Producer"`
	EndpointConsumer rocketmq.ConsumerConfig `conf:"consumer" gen:"Consumer"`
}

func (c Config) String() string {
	return ""
}
