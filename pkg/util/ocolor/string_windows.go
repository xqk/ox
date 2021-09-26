// +build windows

package ocolor

import (
	"fmt"
	"math/rand"
	"strconv"
)

var _ = RandomColor()

//
// RandomColor
// @Description: 生成随机颜色
// @return string
//
func RandomColor() string {
	return fmt.Sprintf("#%s", strconv.FormatInt(int64(rand.Intn(16777216)), 16))
}

//
// Yellow
// @Description:
// @param msg
// @param arg
// @return string
//
func Yellow(msg string, arg ...interface{}) string {
	return sprint(msg, arg...)
}

//
// Red
// @Description:
// @param msg
// @param arg
// @return string
//
func Red(msg string, arg ...interface{}) string {
	return sprint(msg, arg...)
}

//
// Blue
// @Description:
// @param msg
// @param arg
// @return string
//
func Blue(msg string, arg ...interface{}) string {
	return sprint(msg, arg...)
}

//
// Green
// @Description:
// @param msg
// @param arg
// @return string
//
func Green(msg string, arg ...interface{}) string {
	return sprint(msg, arg...)
}

//
// sprint
// @Description:
// @param msg
// @param arg
// @return string
//
func sprint(msg string, arg ...interface{}) string {
	if arg != nil {
		return fmt.Sprintf("%s %+v\n", msg, arrToTransform(arg))
	} else {
		return fmt.Sprintf("%s", msg)
	}
}
