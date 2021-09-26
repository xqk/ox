package etcdv3

import (
	"ox/pkg/registry"
)

func init() {
	registry.RegisterBuilder("etcdv3", func(confKey string) registry.Registry {
		return RawConfig(confKey).MustBuild()
	})
}
