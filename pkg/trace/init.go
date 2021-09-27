package trace

import (
	"log"

	"github.com/xqk/ox/pkg/conf"
	"github.com/xqk/ox/pkg/trace/jaeger"
)

func init() {
	// 加载完配置，初始化sentinel
	conf.OnLoaded(func(c *conf.Configuration) {
		log.Print("hook config, init sentinel rules")
		if conf.Get("ox.trace.jaeger") != nil {
			var config = jaeger.RawConfig("ox.trace.jaeger")
			SetGlobalTracer(config.Build())
		}
	})
}
