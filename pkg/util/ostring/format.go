package ostring

import "fmt"

// Formatter 格式对象
type Formatter string

//
// Format
// @Description: 格式化
// @receiver fm
// @param args
// @return string
//
func (fm Formatter) Format(args ...interface{}) string {
	return fmt.Sprintf(string(fm), args...)
}
