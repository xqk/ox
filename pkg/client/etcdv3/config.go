package etcdv3

import (
	"github.com/xqk/ox/pkg/olog"
	"github.com/xqk/ox/pkg/util/otime"
	"time"

	"github.com/xqk/ox/pkg/conf"
	"github.com/xqk/ox/pkg/constant"
	"github.com/xqk/ox/pkg/ecode"
	"github.com/xqk/ox/pkg/flag"
)

var ConfigPrefix = constant.ConfigPrefix + ".etcdv3"

// Config ...
type (
	Config struct {
		Endpoints []string `json:"endpoints"`
		CertFile  string   `json:"certFile"`
		KeyFile   string   `json:"keyFile"`
		CaCert    string   `json:"caCert"`
		BasicAuth bool     `json:"basicAuth"`
		UserName  string   `json:"userName"`
		Password  string   `json:"-"`
		// 连接超时时间
		ConnectTimeout time.Duration `json:"connectTimeout"`
		Secure         bool          `json:"secure"`
		// 自动同步member list的间隔
		AutoSyncInterval time.Duration `json:"autoAsyncInterval"`
		TTL              int           // 单位：s
		logger           *olog.Logger
	}
)

func (config *Config) BindFlags(fs *flag.FlagSet) {
	fs.BoolVar(&config.Secure, "insecure-etcd", true, "--insecure-etcd=true")
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		BasicAuth:      false,
		ConnectTimeout: otime.Duration("5s"),
		Secure:         false,
		logger:         olog.OxLogger.With(olog.FieldMod("client.etcd")),
	}
}

// StdConfig ...
func StdConfig(name string) *Config {
	return RawConfig(ConfigPrefix + "." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, config); err != nil {
		config.logger.Panic("client etcd parse config panic", olog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), olog.FieldErr(err), olog.FieldKey(key), olog.FieldValueAny(config))
	}
	return config
}

// WithLogger ...
func (config *Config) WithLogger(logger *olog.Logger) *Config {
	config.logger = logger
	return config
}

// Build ...
func (config *Config) Build() (*Client, error) {
	return newClient(config)
}

func (config *Config) MustBuild() *Client {
	client, err := config.Build()
	if err != nil {
		olog.Panicf("build etcd client failed: %v", err)
	}
	return client
}
