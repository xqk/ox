package ogrpc

import (
	"fmt"
	"google.golang.org/grpc"
	"ox/pkg/conf"
	"ox/pkg/constant"
	"ox/pkg/ecode"
	"ox/pkg/flag"
	"ox/pkg/olog"
)

// Config ...
type Config struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Deployment string `json:"deployment"`
	// Network network type, tcp4 by default
	Network string `json:"network" toml:"network"`
	// DisableTrace disbale Trace Interceptor, false by default
	DisableTrace bool
	// DisableMetric disable Metric Interceptor, false by default
	DisableMetric bool
	// SlowQueryThresholdInMilli, request will be colored if cost over this threshold value
	SlowQueryThresholdInMilli int64
	// ServiceAddress service address in registry info, default to 'Host:Port'
	ServiceAddress string

	Labels map[string]string `json:"labels"`

	serverOptions      []grpc.ServerOption
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor

	logger *olog.Logger
}

// StdConfig represents Standard gRPC Server config
// which will parse config by conf package,
// panic if no config key found in conf
func StdConfig(name string) *Config {
	return RawConfig("ox.server." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil {
		config.logger.Panic("grpc server parse config panic",
			olog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr),
			olog.FieldErr(err), olog.FieldKey(key),
			olog.FieldValueAny(config),
		)
	}
	return config
}

// DefaultConfig represents default config
// User should construct config base on DefaultConfig
func DefaultConfig() *Config {
	return &Config{
		Network:                   "tcp4",
		Host:                      flag.String("host"),
		Port:                      9092,
		Deployment:                constant.DefaultDeployment,
		DisableMetric:             false,
		DisableTrace:              false,
		SlowQueryThresholdInMilli: 500,
		logger:                    olog.OxLogger.With(olog.FieldMod("server.grpc")),
		serverOptions:             []grpc.ServerOption{},
		streamInterceptors:        []grpc.StreamServerInterceptor{},
		unaryInterceptors:         []grpc.UnaryServerInterceptor{},
	}
}

// WithServerOption inject server option to grpc server
// User should not inject interceptor option, which is recommend by WithStreamInterceptor
// and WithUnaryInterceptor
func (config *Config) WithServerOption(options ...grpc.ServerOption) *Config {
	if config.serverOptions == nil {
		config.serverOptions = make([]grpc.ServerOption, 0)
	}
	config.serverOptions = append(config.serverOptions, options...)
	return config
}

// WithStreamInterceptor inject stream interceptors to server option
func (config *Config) WithStreamInterceptor(intes ...grpc.StreamServerInterceptor) *Config {
	if config.streamInterceptors == nil {
		config.streamInterceptors = make([]grpc.StreamServerInterceptor, 0)
	}

	config.streamInterceptors = append(config.streamInterceptors, intes...)
	return config
}

// WithUnaryInterceptor inject unary interceptors to server option
func (config *Config) WithUnaryInterceptor(intes ...grpc.UnaryServerInterceptor) *Config {
	if config.unaryInterceptors == nil {
		config.unaryInterceptors = make([]grpc.UnaryServerInterceptor, 0)
	}

	config.unaryInterceptors = append(config.unaryInterceptors, intes...)
	return config
}

func (config *Config) MustBuild() *Server {
	server, err := config.Build()
	if err != nil {
		olog.Panicf("build ogrpc server: %v", err)
	}
	return server
}

// Build ...
func (config *Config) Build() (*Server, error) {
	if !config.DisableTrace {
		config.unaryInterceptors = append(config.unaryInterceptors, traceUnaryServerInterceptor)
		config.streamInterceptors = append(config.streamInterceptors, traceStreamServerInterceptor)
	}

	if !config.DisableMetric {
		config.unaryInterceptors = append(config.unaryInterceptors, prometheusUnaryServerInterceptor)
		config.streamInterceptors = append(config.streamInterceptors, prometheusStreamServerInterceptor)
	}

	return newServer(config)
}

// WithLogger ...
func (config *Config) WithLogger(logger *olog.Logger) *Config {
	config.logger = logger
	return config
}

// Address ...
func (config Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
