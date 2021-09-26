package etcdv3

import (
	"ox/pkg/olog"
	"time"

	"ox/pkg/client/etcdv3"
	"ox/pkg/conf"
	"ox/pkg/ecode"
	"ox/pkg/registry"
)

// StdConfig ...
func StdConfig(name string) *Config {
	return RawConfig("ox.registry." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	// 解析最外层配置
	if err := conf.UnmarshalKey(key, &config); err != nil {
		olog.Panic("unmarshal key", olog.FieldMod("registry.etcd"), olog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), olog.FieldErr(err), olog.String("key", key), olog.Any("config", config))
	}
	// 解析嵌套配置
	if err := conf.UnmarshalKey(key, &config.Config); err != nil {
		olog.Panic("unmarshal key", olog.FieldMod("registry.etcd"), olog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), olog.FieldErr(err), olog.String("key", key), olog.Any("config", config))
	}
	return config
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Config:      etcdv3.DefaultConfig(),
		ReadTimeout: time.Second * 3,
		Prefix:      "ox",
		logger:      olog.OxLogger,
		ServiceTTL:  0,
	}
}

// Config ...
type Config struct {
	*etcdv3.Config
	ReadTimeout time.Duration
	ConfigKey   string
	Prefix      string
	ServiceTTL  time.Duration
	logger      *olog.Logger
}

// Build ...
func (config Config) Build() (registry.Registry, error) {
	if config.ConfigKey != "" {
		config.Config = etcdv3.RawConfig(config.ConfigKey)
	}
	return newETCDRegistry(&config)
}

func (config Config) MustBuild() registry.Registry {
	reg, err := config.Build()
	if err != nil {
		olog.Panicf("build registry failed: %v", err)
	}
	return reg
}
