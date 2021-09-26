package governor

import (
	"fmt"
	"ox/pkg/olog"

	"ox/pkg/conf"
	"ox/pkg/util/onet"
)

//ModName ..
const ModName = "govern"

// Config ...
type Config struct {
	Host    string
	Port    int
	Network string `json:"network" toml:"network"`
	logger  *olog.Logger
	Enable  bool

	// ServiceAddress service address in registry info, default to 'Host:Port'
	ServiceAddress string
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
	if conf.Get(key) == nil {
		return config
	}
	if err := conf.UnmarshalKey(key, &config); err != nil {
		config.logger.Panic("govern server parse config panic",
			olog.FieldErr(err), olog.FieldKey(key),
			olog.FieldValueAny(config),
		)
	}
	return config
}

// DefaultConfig represents default config
// User should construct config base on DefaultConfig
func DefaultConfig() *Config {
	host, port, err := onet.GetLocalMainIP()
	if err != nil {
		host = "localhost"
	}

	return &Config{
		Enable:  true,
		Host:    host,
		Network: "tcp4",
		Port:    port,
		logger:  olog.OxLogger.With(olog.FieldMod(ModName)),
	}
}

// Build ...
func (config *Config) Build() *Server {
	return newServer(config)
}

// Address ...
func (config Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
