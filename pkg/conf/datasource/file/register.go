package file

import (
	"github.com/xqk/ox/pkg/conf"
	"github.com/xqk/ox/pkg/flag"
	"github.com/xqk/ox/pkg/olog"
)

// DataSourceFile defines file scheme
const DataSourceFile = "file"

func init() {
	conf.Register(DataSourceFile, func() conf.DataSource {
		var (
			watchConfig = flag.Bool("watch")
			configAddr  = flag.String("config")
		)
		if configAddr == "" {
			olog.Panic("new file dataSource, configAddr is empty")
			return nil
		}
		return NewDataSource(configAddr, watchConfig)
	})
}
