package xcolor

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
	return sprint(YellowColor, msg, arg...)
}

//
// Red
// @Description:
// @param msg
// @param arg
// @return string
//
func Red(msg string, arg ...interface{}) string {
	return sprint(RedColor, msg, arg...)
}

//
// Blue
// @Description:
// @param msg
// @param arg
// @return string
//
func Blue(msg string, arg ...interface{}) string {
	return sprint(BlueColor, msg, arg...)
}

//
// Green
// @Description:
// @param msg
// @param arg
// @return string
//
func Green(msg string, arg ...interface{}) string {
	return sprint(GreenColor, msg, arg...)
}

//
// sprint
// @Description:
// @param colorValue
// @param msg
// @param arg
// @return string
//
func sprint(colorValue int, msg string, arg ...interface{}) string {
	if arg != nil {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m %+v", colorValue, msg, arrToTransform(arg))
	} else {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorValue, msg)
	}
}
