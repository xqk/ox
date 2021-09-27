package ogo

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/codegangsta/inject"
	"github.com/xqk/ox/pkg/olog"
)

//
// Serial
// @Description: 串行
// @param fns
// @return func()
//
func Serial(fns ...func()) func() {
	return func() {
		for _, fn := range fns {
			fn()
		}
	}
}

//
// Parallel
// @Description: 并发执行
// @param fns
// @return func()
//
func Parallel(fns ...func()) func() {
	var wg sync.WaitGroup
	return func() {
		wg.Add(len(fns))
		for _, fn := range fns {
			go try2(fn, wg.Done)
		}
		wg.Wait()
	}
}

//
// RestrictParallel
// @Description: 并发,最大并发量restrict
// @param restrict
// @param fns
// @return func()
//
func RestrictParallel(restrict int, fns ...func()) func() {
	var channel = make(chan struct{}, restrict)
	return func() {
		var wg sync.WaitGroup
		for _, fn := range fns {
			wg.Add(1)
			go func(fn func()) {
				defer wg.Done()
				channel <- struct{}{}
				try2(fn, nil)
				<-channel
			}(fn)
		}
		wg.Wait()
		close(channel)
	}
}

//
// GoDirect
// @Description:
// @param fn
// @param args
//
func GoDirect(fn interface{}, args ...interface{}) {
	var inj = inject.New()
	for _, arg := range args {
		inj.Map(arg)
	}

	_, file, line, _ := runtime.Caller(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				_logger.Error("recover", olog.Any("err", err), olog.String("line", fmt.Sprintf("%s:%d", file, line)))
			}
		}()
		// 忽略返回值, goroutine执行的返回值通常都会忽略掉
		_, err := inj.Invoke(fn)
		if err != nil {
			_logger.Error("inject", olog.Any("err", err), olog.String("line", fmt.Sprintf("%s:%d", file, line)))
			return
		}
	}()
}

//
// Go
// @Description: goroutine
// @param fn
//
func Go(fn func()) {
	go try2(fn, nil)
}

//
// DelayGo
// @Description: goroutine
// @param delay
// @param fn
//
func DelayGo(delay time.Duration, fn func()) {
	_, file, line, _ := runtime.Caller(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				_logger.Error("recover", olog.Any("err", err), olog.String("line", fmt.Sprintf("%s:%d", file, line)))
			}
		}()
		time.Sleep(delay)
		fn()
	}()
}

//
// SafeGo
// @Description: safe go
// @param fn
// @param rec
//
func SafeGo(fn func(), rec func(error)) {
	go func() {
		err := try2(fn, nil)
		if err != nil {
			rec(err)
		}
	}()
}
