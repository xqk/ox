package oecho

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/xqk/ox/pkg/conf"
	"github.com/xqk/ox/pkg/constant"
	"github.com/xqk/ox/pkg/ecode"
	"github.com/xqk/ox/pkg/flag"
	"github.com/xqk/ox/pkg/olog"
)

//ModName named a mod
const ModName = "server.echo"

//Config HTTP config
type Config struct {
	Host          string
	Port          int
	Deployment    string
	Debug         bool
	DisableMetric bool
	DisableTrace  bool
	// ServiceAddress service address in registry info, default to 'Host:Port'
	ServiceAddress string

	SlowQueryThresholdInMilli int64

	logger *olog.Logger
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Host:                      flag.String("host"),
		Port:                      9091,
		Debug:                     false,
		Deployment:                constant.DefaultDeployment,
		SlowQueryThresholdInMilli: 500, // 500ms
		logger:                    olog.OxLogger.With(olog.FieldMod(ModName)),
	}
}

// StdConfig Ox Standard HTTP Server config
func StdConfig(name string) *Config {
	return RawConfig("ox.server." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil &&
		errors.Cause(err) != conf.ErrInvalidKey {
		config.logger.Panic("http server parse config panic", olog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), olog.FieldErr(err), olog.FieldKey(key), olog.FieldValueAny(config))
	}
	return config
}

// WithLogger ...
func (config *Config) WithLogger(logger *olog.Logger) *Config {
	config.logger = logger
	return config
}

// WithHost ...
func (config *Config) WithHost(host string) *Config {
	config.Host = host
	return config
}

// WithPort ...
func (config *Config) WithPort(port int) *Config {
	config.Port = port
	return config
}

func (config *Config) MustBuild() *Server {
	server, err := config.Build()
	if err != nil {
		olog.Panicf("build echo server failed: %v", err)
	}
	return server
}

// Build create server instance, then initialize it with necessary interceptor
func (config *Config) Build() (*Server, error) {
	server, err := newServer(config)
	if err != nil {
		return nil, err
	}
	server.Use(recoverMiddleware(config.logger, config.SlowQueryThresholdInMilli))

	if !config.DisableMetric {
		server.Use(metricServerInterceptor())
	}

	if !config.DisableTrace {
		server.Use(traceServerInterceptor())
	}
	return server, nil
}

// Address ...
func (config *Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
