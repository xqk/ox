package ostring

import (
	"reflect"
	"runtime"
)

//
// FunctionName
// @Description:方法名
// @param i
// @return string
//
func FunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

//
// ObjectName
// @Description:对象名
// @param i
// @return string
//
func ObjectName(i interface{}) string {
	typ := reflect.TypeOf(i)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	return typ.PkgPath() + "." + typ.Name()
}

//
// CallerName
// @Description:调用者名称
// @param skip
// @return string
//
func CallerName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	return runtime.FuncForPC(pc).Name()
}
