package http

import (
	"github.com/xqk/ox/pkg/conf"
	"github.com/xqk/ox/pkg/flag"
	"github.com/xqk/ox/pkg/olog"
)

// Defines http/https scheme
const (
	DataSourceHttp  = "http"
	DataSourceHttps = "https"
)

func init() {
	dataSourceCreator := func() conf.DataSource {
		var (
			watchConfig = flag.Bool("watch")
			configAddr  = flag.String("config")
		)
		if configAddr == "" {
			olog.Panic("new http dataSource, configAddr is empty")
			return nil
		}
		return NewDataSource(configAddr, watchConfig)
	}
	conf.Register(DataSourceHttp, dataSourceCreator)
	conf.Register(DataSourceHttps, dataSourceCreator)
}
