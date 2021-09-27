package ogo

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"github.com/xqk/ox/pkg/olog"
	"github.com/xqk/ox/pkg/util/ostring"
)

var (
	_logger = olog.OxLogger.With(zap.String("mod", "xgo"))
)

func try(fn func() error, cleaner func()) (ret error) {
	if cleaner != nil {
		defer cleaner()
	}
	defer func() {
		if err := recover(); err != nil {
			_, file, line, _ := runtime.Caller(2)
			_logger.Error("recover", zap.Any("err", err), zap.String("line", fmt.Sprintf("%s:%d", file, line)))
			if _, ok := err.(error); ok {
				ret = err.(error)
			} else {
				ret = fmt.Errorf("%+v", err)
			}
			ret = errors.Wrap(ret, fmt.Sprintf("%s:%d", ostring.FunctionName(fn), line))
		}
	}()
	return fn()
}

func try2(fn func(), cleaner func()) (ret error) {
	if cleaner != nil {
		defer cleaner()
	}
	defer func() {
		_, file, line, _ := runtime.Caller(5)
		if err := recover(); err != nil {
			_logger.Error("recover", zap.Any("err", err), zap.String("line", fmt.Sprintf("%s:%d", file, line)))
			if _, ok := err.(error); ok {
				ret = err.(error)
			} else {
				ret = fmt.Errorf("%+v", err)
			}
		}
	}()
	fn()
	return nil
}
