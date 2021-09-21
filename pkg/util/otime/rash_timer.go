package otime

import (
	"sync"
	"time"
)

var defaultWheel *rashTimer

func init() {
	defaultWheel = NewRashTimer(500 * time.Millisecond)
}

type Timer struct {
	C <-chan time.Time
	r *timer
}

//
// After
// @Description:
// @param d
// @return <-chan
//
func After(d time.Duration) <-chan time.Time {
	return defaultWheel.After(d)
}

//
// Sleep
// @Description:
// @param d
//
func Sleep(d time.Duration) {
	defaultWheel.Sleep(d)
}

//
// AfterFunc
// @Description:
// @param d
// @param f
// @return *Timer
//
func AfterFunc(d time.Duration, f func()) *Timer {
	return defaultWheel.AfterFunc(d, f)
}

//
// NewTimer
// @Description:
// @param d
// @return *Timer
//
func NewTimer(d time.Duration) *Timer {
	return defaultWheel.NewTimer(d)
}

//
// Reset
// @Description:
// @receiver t
// @param d
//
func (t *Timer) Reset(d time.Duration) {
	t.r.w.resetTimer(t.r, d, 0)
}

//
// Stop
// @Description:
// @receiver t
//
func (t *Timer) Stop() {
	t.r.w.delTimer(t.r)
}

//
// Ticker
// @Description: 定时器
//
type Ticker struct {
	C <-chan time.Time
	r *timer
}

//
// NewTicker
// @Description: 新建定时器
// @param d
// @return *Ticker
//
func NewTicker(d time.Duration) *Ticker {
	return defaultWheel.NewTicker(d)
}

//
// TickFunc
// @Description: 定时方法
// @param d
// @param f
// @return *Ticker
//
func TickFunc(d time.Duration, f func()) *Ticker {
	return defaultWheel.TickFunc(d, f)
}

//
// Tick
// @Description: 定时
// @param d
// @return <-chan
//
func Tick(d time.Duration) <-chan time.Time {
	return defaultWheel.Tick(d)
}

//
// Stop
// @Description: 停止
// @receiver t
//
func (t *Ticker) Stop() {
	t.r.w.delTimer(t.r)
}

//
// Reset
// @Description: 重启
// @receiver t
// @param d
//
func (t *Ticker) Reset(d time.Duration) {
	t.r.w.resetTimer(t.r, d, d)
}

const (
	tvn_bits uint64 = 6
	tvr_bits uint64 = 8
	tvn_size uint64 = 64  // 1 << tvn_bits
	tvr_size uint64 = 256 // 1 << tvr_bits

	tvn_mask uint64 = 63  // tvn_size - 1
	tvr_mask uint64 = 255 // tvr_size -1
)

const (
	defaultTimerSize = 128
)

type timer struct {
	expires uint64
	period  uint64

	f   func(time.Time, interface{})
	arg interface{}

	w *rashTimer

	vec   []*timer
	index int
}

//
// rashTimer
// @Description: 低精度timer
//
type rashTimer struct {
	sync.Mutex

	jiffies uint64

	tv1 [][]*timer
	tv2 [][]*timer
	tv3 [][]*timer
	tv4 [][]*timer
	tv5 [][]*timer

	tick time.Duration

	quit chan struct{}
}

// NewRashTimer is the time for a jiffies
func NewRashTimer(tick time.Duration) *rashTimer {
	w := new(rashTimer)

	w.quit = make(chan struct{})

	f := func(size int) [][]*timer {
		tv := make([][]*timer, size)
		for i := range tv {
			tv[i] = make([]*timer, 0, defaultTimerSize)
		}

		return tv
	}

	w.tv1 = f(int(tvr_size))
	w.tv2 = f(int(tvn_size))
	w.tv3 = f(int(tvn_size))
	w.tv4 = f(int(tvn_size))
	w.tv5 = f(int(tvn_size))

	w.jiffies = 0
	w.tick = tick

	go w.run()
	return w
}

//
// addTimerInternal
// @Description: 添加定时
// @receiver w
// @param t
//
func (w *rashTimer) addTimerInternal(t *timer) {
	expires := t.expires
	idx := t.expires - w.jiffies

	var tv [][]*timer
	var i uint64

	if idx < tvr_size {
		i = expires & tvr_mask
		tv = w.tv1
	} else if idx < (1 << (tvr_bits + tvn_bits)) {
		i = (expires >> tvr_bits) & tvn_mask
		tv = w.tv2
	} else if idx < (1 << (tvr_bits + 2*tvn_bits)) {
		i = (expires >> (tvr_bits + tvn_bits)) & tvn_mask
		tv = w.tv3
	} else if idx < (1 << (tvr_bits + 3*tvn_bits)) {
		i = (expires >> (tvr_bits + 2*tvn_bits)) & tvn_mask
		tv = w.tv4
	} else if int64(idx) < 0 {
		i = w.jiffies & tvr_mask
		tv = w.tv1
	} else {
		if idx > 0x00000000ffffffff {
			idx = 0x00000000ffffffff

			expires = idx + w.jiffies
		}

		i = (expires >> (tvr_bits + 3*tvn_bits)) & tvn_mask
		tv = w.tv5
	}

	tv[i] = append(tv[i], t)

	t.vec = tv[i]
	t.index = len(tv[i]) - 1
}

//
// cascade
// @Description:串联
// @receiver w
// @param tv
// @param index
// @return int
//
func (w *rashTimer) cascade(tv [][]*timer, index int) int {
	vec := tv[index]
	tv[index] = vec[0:0:defaultTimerSize]

	for _, t := range vec {
		w.addTimerInternal(t)
	}

	return index
}

//
// getIndex
// @Description:获取索引
// @receiver w
// @param n
// @return int
//
func (w *rashTimer) getIndex(n int) int {
	return int((w.jiffies >> (tvr_bits + uint64(n)*tvn_bits)) & tvn_mask)
}

//
// onTick
// @Description:开启定时
// @receiver w
//
func (w *rashTimer) onTick() {
	w.Lock()

	index := int(w.jiffies & tvr_mask)

	if index == 0 && (w.cascade(w.tv2, w.getIndex(0))) == 0 &&
		(w.cascade(w.tv3, w.getIndex(1))) == 0 &&
		(w.cascade(w.tv4, w.getIndex(2))) == 0 &&
		(w.cascade(w.tv5, w.getIndex(3)) == 0) {

	}

	w.jiffies++

	vec := w.tv1[index]
	w.tv1[index] = vec[0:0:defaultTimerSize]

	w.Unlock()

	f := func(vec []*timer) {
		now := time.Now()
		for _, t := range vec {
			if t == nil {
				continue
			}

			t.f(now, t.arg)

			if t.period > 0 {
				t.expires = t.period + w.jiffies
				w.addTimer(t)
			}
		}
	}

	if len(vec) > 0 {
		go f(vec)
	}
}

//
// addTimer
// @Description:添加定时时间
// @receiver w
// @param t
//
func (w *rashTimer) addTimer(t *timer) {
	w.Lock()
	w.addTimerInternal(t)
	w.Unlock()
}

//
// delTimer
// @Description:删除定时时间
// @receiver w
// @param t
//
func (w *rashTimer) delTimer(t *timer) {
	w.Lock()
	vec := t.vec
	index := t.index

	if len(vec) > index && vec[index] == t {
		vec[index] = nil
	}

	w.Unlock()
}

//
// resetTimer
// @Description:重启定时时间
// @receiver w
// @param t
// @param when
// @param period
//
func (w *rashTimer) resetTimer(t *timer, when time.Duration, period time.Duration) {
	w.delTimer(t)

	t.expires = w.jiffies + uint64(when/w.tick)
	t.period = uint64(period / w.tick)

	w.addTimer(t)
}

//
// newTimer
// @Description: 新建定时器
// @receiver w
// @param when
// @param period
// @param f
// @param arg
// @return *timer
//
func (w *rashTimer) newTimer(when time.Duration, period time.Duration,
	f func(time.Time, interface{}), arg interface{}) *timer {
	t := new(timer)

	t.expires = w.jiffies + uint64(when/w.tick)
	t.period = uint64(period / w.tick)

	t.f = f
	t.arg = arg

	t.w = w

	return t
}

//
// run
// @Description:启动
// @receiver w
//
func (w *rashTimer) run() {
	ticker := time.NewTicker(w.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.onTick()
		case <-w.quit:
			return
		}
	}
}

//
// Stop
// @Description:停止
// @receiver w
//
func (w *rashTimer) Stop() {
	close(w.quit)
}

func sendTime(t time.Time, arg interface{}) {
	select {
	case arg.(chan time.Time) <- t:
	default:
	}
}

//
// goFunc
// @Description:
// @param t
// @param arg
//
func goFunc(t time.Time, arg interface{}) {
	go arg.(func())()
}

//
// After
// @Description:
// @receiver w
// @param d
// @return <-chan
//
func (w *rashTimer) After(d time.Duration) <-chan time.Time {
	return w.NewTimer(d).C
}

//
// Sleep
// @Description:睡眠
// @receiver w
// @param d
//
func (w *rashTimer) Sleep(d time.Duration) {
	<-w.NewTimer(d).C
}

//
// Tick
// @Description:
// @receiver w
// @param d
// @return <-chan
//
func (w *rashTimer) Tick(d time.Duration) <-chan time.Time {
	return w.NewTicker(d).C
}

//
// TickFunc
// @Description:
// @receiver w
// @param d
// @param f
// @return *Ticker
//
func (w *rashTimer) TickFunc(d time.Duration, f func()) *Ticker {
	t := &Ticker{
		r: w.newTimer(d, d, goFunc, f),
	}

	w.addTimer(t.r)

	return t

}

//
// AfterFunc
// @Description:
// @receiver w
// @param d
// @param f
// @return *Timer
//
func (w *rashTimer) AfterFunc(d time.Duration, f func()) *Timer {
	t := &Timer{
		r: w.newTimer(d, 0, goFunc, f),
	}

	w.addTimer(t.r)

	return t
}

//
// NewTimer
// @Description:
// @receiver w
// @param d
// @return *Timer
//
func (w *rashTimer) NewTimer(d time.Duration) *Timer {
	c := make(chan time.Time, 1)
	t := &Timer{
		C: c,
		r: w.newTimer(d, 0, sendTime, c),
	}

	w.addTimer(t.r)

	return t
}

//
// NewTicker
// @Description:
// @receiver w
// @param d
// @return *Ticker
//
func (w *rashTimer) NewTicker(d time.Duration) *Ticker {
	c := make(chan time.Time, 1)
	t := &Ticker{
		C: c,
		r: w.newTimer(d, d, sendTime, c),
	}

	w.addTimer(t.r)

	return t
}
