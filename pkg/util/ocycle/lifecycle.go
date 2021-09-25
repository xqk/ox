package ocycle

import (
	"sync"
	"sync/atomic"
)

//
// Cycle
// @Description:
//
type Cycle struct {
	mu      *sync.Mutex
	wg      *sync.WaitGroup
	done    chan struct{}
	quit    chan error
	closing uint32
	waiting uint32
	// works []func() error
}

//
// NewCycle
// @Description: 新增循环周期
// @return *Cycle
//
func NewCycle() *Cycle {
	return &Cycle{
		mu:      &sync.Mutex{},
		wg:      &sync.WaitGroup{},
		done:    make(chan struct{}),
		quit:    make(chan error),
		closing: 0,
		waiting: 0,
	}
}

//
// Run
// @Description: 启动一个goroutine
// @receiver c
// @param fn
//
func (c *Cycle) Run(fn func() error) {
	c.mu.Lock()
	//todo add check options panic before waiting
	defer c.mu.Unlock()
	c.wg.Add(1)
	go func(c *Cycle) {
		defer c.wg.Done()
		if err := fn(); err != nil {
			c.quit <- err
		}
	}(c)
}

//
// Done
// @Description: 阻塞并返回一个chan错误
// @receiver c
// @return <-chan
//
func (c *Cycle) Done() <-chan struct{} {
	if atomic.CompareAndSwapUint32(&c.waiting, 0, 1) {
		go func(c *Cycle) {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.wg.Wait()
			close(c.done)
		}(c)
	}
	return c.done
}

//
// DoneAndClose
// @Description:
// @receiver c
//
func (c *Cycle) DoneAndClose() {
	<-c.Done()
	c.Close()
}

//
// Close
// @Description:
// @receiver c
//
func (c *Cycle) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if atomic.CompareAndSwapUint32(&c.closing, 0, 1) {
		close(c.quit)
	}
}

//
// Wait
// @Description:阻塞生命周期
// @receiver c
// @return <-chan
//
func (c *Cycle) Wait() <-chan error {
	return c.quit
}
