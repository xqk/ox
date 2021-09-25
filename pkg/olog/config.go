package olog

import (
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"ox/pkg"
	"ox/pkg/conf"
	"ox/pkg/constant"
)

func init() {
	conf.OnLoaded(func(c *conf.Configuration) {
		log.Print("hook config, init loggers")
		if c.Get(ConfigEntry("default")) != nil {
			log.Printf("reload default logger with configKey: %s", ConfigEntry("default"))
			DefaultLogger = RawConfig(constant.ConfigPrefix + ".logger.default").Build()
		}
		DefaultLogger.AutoLevel(constant.ConfigPrefix + ".logger.default")

		if c.Get(constant.ConfigPrefix+".logger.ox") != nil {
			log.Printf("reload default logger with configKey: %s", ConfigEntry("ox"))
			OxLogger = RawConfig(constant.ConfigPrefix + ".logger.ox").Build()
		}
		OxLogger.AutoLevel(constant.ConfigPrefix + ".logger.ox")
	})
}

var ConfigPrefix = constant.ConfigPrefix + ".logger"

// Config ...
type Config struct {
	// Dir 日志输出目录
	Dir string
	// Name 日志文件名称
	Name string
	// Level 日志初始等级
	Level string
	// 日志初始化字段
	Fields []zap.Field
	// 是否添加调用者信息
	AddCaller bool
	// 日志前缀
	Prefix string
	// 日志输出文件最大长度，超过改值则截断
	MaxSize   int
	MaxAge    int
	MaxBackup int
	// 日志磁盘刷盘间隔
	Interval      time.Duration
	CallerSkip    int
	Async         bool
	Queue         bool
	QueueSleep    time.Duration
	Core          zapcore.Core
	Debug         bool
	EncoderConfig *zapcore.EncoderConfig
	configKey     string
}

// Filename ...
func (config *Config) Filename() string {
	return fmt.Sprintf("%s/%s", config.Dir, config.Name)
}

func ConfigEntry(name string) string {
	return ConfigPrefix + "." + name
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil {
		panic(err)
	}
	config.configKey = key
	return config
}

// StdConfig Ox Standard logger config
func StdConfig(name string) *Config {
	return RawConfig(ConfigPrefix + "." + name)
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Name:          "default.log",
		Dir:           pkg.LogDir(),
		Level:         "info",
		MaxSize:       500, // 500M
		MaxAge:        1,   // 1 day
		MaxBackup:     10,  // 10 backup
		Interval:      24 * time.Hour,
		CallerSkip:    1,
		AddCaller:     true,
		Async:         true,
		Queue:         false,
		QueueSleep:    100 * time.Millisecond,
		EncoderConfig: DefaultZapConfig(),
		Fields: []zap.Field{
			String("aid", pkg.AppID()),
			String("iid", pkg.AppInstance()),
		},
	}
}

// Build ...
func (config Config) Build() *Logger {
	if config.EncoderConfig == nil {
		config.EncoderConfig = DefaultZapConfig()
	}
	if config.Debug {
		config.EncoderConfig.EncodeLevel = DebugEncodeLevel
	}
	logger := newLogger(&config)
	if config.configKey != "" {
		logger.AutoLevel(config.configKey + ".level")
	}
	return logger
}
