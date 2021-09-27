package sentinel

import (
	"log"
	"github.com/xqk/ox/pkg/conf"
)

func init() {
	// 加载完配置，初始化sentinel
	conf.OnLoaded(func(c *conf.Configuration) {
		log.Print("hook config, init sentinel rules")
		var config = DefaultConfig()
		if err := conf.UnmarshalKey("sentinel", &config, conf.BuildinModule("reliability")); err != nil {
			log.Printf("read sentinel config failed %v", err)
			return
		}

		// initialize global sentinel
		config.Build()
	})
}
