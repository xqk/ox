package otime

import (
	"sync/atomic"
	"time"
)

var nowInMs = uint64(0)

//
// StartTimeTicker
// @Description: 开始时间定时器
//
func StartTimeTicker() {
	atomic.StoreUint64(&nowInMs, uint64(time.Now().UnixNano())/UnixTimeUnitOffset)
	go func() {
		for {
			now := uint64(time.Now().UnixNano()) / UnixTimeUnitOffset
			atomic.StoreUint64(&nowInMs, now)
			time.Sleep(time.Millisecond)
		}
	}()
}

//
// CurrentTimeMillsWithTicker
// @Description: 当前时间的定时器
// @return uint64
//
func CurrentTimeMillsWithTicker() uint64 {
	return atomic.LoadUint64(&nowInMs)
}
