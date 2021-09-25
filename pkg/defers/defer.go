package defers

import (
	"ox/pkg/util/odefer"
)

var (
	globalDefers = odefer.NewStack()
)

//
// Register
// @Description: 注册一个defer函数
// @param fns
//
func Register(fns ...func() error) {
	globalDefers.Push(fns...)
}

//
// Clean
// @Description: 清除
//
func Clean() {
	globalDefers.Clean()
}
