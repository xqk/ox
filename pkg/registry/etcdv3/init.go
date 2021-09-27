package etcdv3

import (
	"github.com/xqk/ox/pkg/registry"
)

func init() {
	registry.RegisterBuilder("etcdv3", func(confKey string) registry.Registry {
		return RawConfig(confKey).MustBuild()
	})
}
