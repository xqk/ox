package etcdv3

import (
	"github.com/xqk/ox/pkg/client/etcdv3"
	"github.com/xqk/ox/pkg/conf"
	"github.com/xqk/ox/pkg/flag"
	"github.com/xqk/ox/pkg/olog"
	"github.com/xqk/ox/pkg/util/onet"
)

// DataSourceEtcdv3 defines etcdv3 scheme
const DataSourceEtcdv3 = "etcdv3"

func init() {
	conf.Register(DataSourceEtcdv3, func() conf.DataSource {
		var (
			configAddr = flag.String("config")
		)
		if configAddr == "" {
			olog.Panic("new apollo dataSource, configAddr is empty")
			return nil
		}
		// configAddr is a string in this format:
		// etcdv3://ip:port?basicAuth=true&username=XXX&password=XXX&key=XXX&certFile=XXX&keyFile=XXX&caCert=XXX&secure=XXX

		urlObj, err := onet.ParseURL(configAddr)
		if err != nil {
			olog.Panic("parse configAddr error", olog.FieldErr(err))
			return nil
		}
		etcdConf := etcdv3.DefaultConfig()
		etcdConf.Endpoints = []string{urlObj.Host}
		etcdConf.BasicAuth = urlObj.QueryBool("basicAuth", false)
		etcdConf.Secure = urlObj.QueryBool("secure", false)
		etcdConf.CertFile = urlObj.Query().Get("certFile")
		etcdConf.KeyFile = urlObj.Query().Get("keyFile")
		etcdConf.CaCert = urlObj.Query().Get("caCert")
		etcdConf.UserName = urlObj.Query().Get("username")
		etcdConf.Password = urlObj.Query().Get("password")
		return NewDataSource(etcdConf.MustBuild(), urlObj.Query().Get("key"))
	})
}
