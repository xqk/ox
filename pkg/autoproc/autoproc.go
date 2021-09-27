package autoproc

import (
	"go.uber.org/automaxprocs/maxprocs"
	"github.com/xqk/ox/pkg/conf"
	"github.com/xqk/ox/pkg/ecode"
	"github.com/xqk/ox/pkg/olog"
	"runtime"
)

func init() {
	// 初始化注册中心
	if _, err := maxprocs.Set(); err != nil {
		olog.Panic("auto max procs", olog.FieldMod(ecode.ModProc), olog.FieldErrKind(ecode.ErrKindAny), olog.FieldErr(err))
	}
	conf.OnLoaded(func(c *conf.Configuration) {
		if maxProcs := conf.GetInt("maxProc"); maxProcs != 0 {
			runtime.GOMAXPROCS(maxProcs)
		}
		olog.Info("auto max procs", olog.FieldMod(ecode.ModProc), olog.Int64("procs", int64(runtime.GOMAXPROCS(-1))))
	})
}
