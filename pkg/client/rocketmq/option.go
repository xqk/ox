package rocketmq

import (
	"github.com/xqk/ox/pkg/conf"
	"github.com/xqk/ox/pkg/olog"
	"github.com/xqk/ox/pkg/util/odebug"
	"strings"
	"time"
)

// Config config...
type Config struct {
	Addresses []string        `json:"addr" toml:"addr"`
	Consumer  *ConsumerConfig `json:"consumer" toml:"consumer"`
	Producer  *ProducerConfig `json:"producer" toml:"producer"`
}

// ConsumerConfig consumer config
type ConsumerConfig struct {
	Name            string        `json:"name" toml:"name"`
	Enable          bool          `json:"enable" toml:"enable"`
	Addr            []string      `json:"addr" toml:"addr"`
	Topic           string        `json:"topic" toml:"topic"`
	Group           string        `json:"group" toml:"group"`
	DialTimeout     time.Duration `json:"dialTimeout" toml:"dialTimeout"`
	RwTimeout       time.Duration `json:"rwTimeout" toml:"rwTimeout"`
	SubExpression   string        `json:"subExpression" toml:"subExpression"`
	Rate            float64       `json:"rate" toml:"rate"`
	Capacity        int64         `json:"capacity" toml:"capacity"`
	WaitMaxDuration time.Duration `json:"waitMaxDuration" toml:"waitMaxDuration"`
	Shadow          Shadow        `json:"shadow" toml:"shadow"`
	Reconsume       int32         `json:"reconsume" toml:"reconsume"`
	AccessKey       string        `json:"accessKey" toml:"accessKey"`
	SecretKey       string        `json:"secretKey" toml:"secretKey"`
}

// ProducerConfig producer config
type ProducerConfig struct {
	Name        string        `json:"name" toml:"name"`
	Addr        []string      `json:"addr" toml:"addr"`
	Topic       string        `json:"topic" toml:"topic"`
	Group       string        `json:"group" toml:"group"`
	Retry       int           `json:"retry" toml:"retry"`
	DialTimeout time.Duration `json:"dialTimeout" toml:"dialTimeout"`
	RwTimeout   time.Duration `json:"rwTimeout" toml:"rwTimeout"`
	Shadow      Shadow        `json:"shadow" toml:"shadow"`
	AccessKey   string        `json:"accessKey" toml:"accessKey"`
	SecretKey   string        `json:"secretKey" toml:"secretKey"`
}

type Shadow struct {
	Mode string `json:"mode" toml:"mode"`
	// mode???????????????????????????topic???????????????
	WitheTopics []string `json:"witheTopics" toml:"witheTopics"`
}

// DefaultConfig ...
func DefaultConfig() Config {
	return Config{
		Addresses: make([]string, 0),
		Producer: &ProducerConfig{
			Retry: 3,
		},
		Consumer: &ConsumerConfig{
			Reconsume: 3,
		},
	}
}

// DefaultConsumerConfig ...
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		DialTimeout: time.Second * 3,
		RwTimeout:   time.Second * 10,
		Reconsume:   3,
	}
}

// DefaultProducerConfig ...
func DefaultProducerConfig() ProducerConfig {
	return ProducerConfig{
		Retry:       3,
		DialTimeout: time.Second * 3,
		RwTimeout:   0,
	}
}

// StdPushConsumerConfig ...
func StdPushConsumerConfig(name string) *ConsumerConfig {

	cc := RawConsumerConfig("ox.rocketmq." + name + ".consumer")
	rc := RawConfig("ox.rocketmq." + name)

	// ??????rocket_client_mq?????????addr????????????shceme
	if len(cc.Addr) == 0 {
		cc.Addr = rc.Addresses
	}
	cc.Name = name
	for ind, addr := range cc.Addr {
		if strings.HasPrefix(addr, "http") {
			cc.Addr[ind] = addr
		} else {
			cc.Addr[ind] = "http://" + addr
		}
	}

	return &cc
}

// StdProducerConfig ...
func StdProducerConfig(name string) *ProducerConfig {
	pc := RawProducerConfig("ox.rocketmq." + name + ".producer")
	rc := RawConfig("ox.rocketmq." + name)
	// ??????rocket_client_mq?????????addr????????????shceme
	if len(pc.Addr) == 0 {
		pc.Addr = rc.Addresses
	}

	pc.Name = name
	for ind, addr := range pc.Addr {
		if strings.HasPrefix(addr, "http") {
			pc.Addr[ind] = addr
		} else {
			pc.Addr[ind] = "http://" + addr
		}
	}
	return &pc
}

// RawConfig ????????????
func RawConfig(key string) Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config, conf.TagName("toml")); err != nil {
		olog.Panic("unmarshal config", olog.String("field", key), olog.Any("ext", config))
	}

	if odebug.IsDevelopmentMode() {
		odebug.PrettyJsonPrint(key, config)
	}
	return config
}

// RawConsumerConfig ????????????
func RawConsumerConfig(key string) ConsumerConfig {
	var config = DefaultConsumerConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil {
		olog.Panic("unmarshal config", olog.String("key", key), olog.Any("config", config))
	}

	return config
}

// RawProducerConfig ...
func RawProducerConfig(key string) ProducerConfig {
	var config = DefaultProducerConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil {
		olog.Panic("unmarshal config", olog.String("key", key))
	}

	return config
}
