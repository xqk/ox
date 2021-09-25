package olog

import (
	"go.uber.org/zap"
)

// DefaultLogger default logger
// Biz Log
// debug=true as default, will be
var DefaultLogger = Config{
	Debug: true,
	Async: true,
}.Build()

// frame logger
var OxLogger = Config{
	Debug: true,
}.Build()

//
// Auto
// @Description:
// @param err
// @return Func
//
func Auto(err error) Func {
	if err != nil {
		return DefaultLogger.With(zap.Any("err", err.Error())).Error
	}

	return DefaultLogger.Info
}

//
// Info
// @Description:
// @param msg
// @param fields
//
func Info(msg string, fields ...Field) {
	DefaultLogger.Info(msg, fields...)
}

//
// Debug
// @Description:
// @param msg
// @param fields
//
func Debug(msg string, fields ...Field) {
	DefaultLogger.Debug(msg, fields...)
}

//
// Warn
// @Description:
// @param msg
// @param fields
//
func Warn(msg string, fields ...Field) {
	DefaultLogger.Warn(msg, fields...)
}

//
// Error
// @Description:
// @param msg
// @param fields
//
func Error(msg string, fields ...Field) {
	DefaultLogger.Error(msg, fields...)
}

//
// Panic
// @Description:
// @param msg
// @param fields
//
func Panic(msg string, fields ...Field) {
	DefaultLogger.Panic(msg, fields...)
}

//
// DPanic
// @Description:
// @param msg
// @param fields
//
func DPanic(msg string, fields ...Field) {
	DefaultLogger.DPanic(msg, fields...)
}

//
// Fatal
// @Description:
// @param msg
// @param fields
//
func Fatal(msg string, fields ...Field) {
	DefaultLogger.Fatal(msg, fields...)
}

//
// Debugw
// @Description:
// @param msg
// @param keysAndValues
//
func Debugw(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Debugw(msg, keysAndValues...)
}

//
// Infow
// @Description:
// @param msg
// @param keysAndValues
//
func Infow(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Infow(msg, keysAndValues...)
}

//
// Warnw
// @Description:
// @param msg
// @param keysAndValues
//
func Warnw(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Warnw(msg, keysAndValues...)
}

//
// Errorw
// @Description:
// @param msg
// @param keysAndValues
//
func Errorw(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Errorw(msg, keysAndValues...)
}

//
// Panicw
// @Description:
// @param msg
// @param keysAndValues
//
func Panicw(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Panicw(msg, keysAndValues...)
}

//
// DPanicw
// @Description:
// @param msg
// @param keysAndValues
//
func DPanicw(msg string, keysAndValues ...interface{}) {
	DefaultLogger.DPanicw(msg, keysAndValues...)
}

//
// Fatalw
// @Description:
// @param msg
// @param keysAndValues
//
func Fatalw(msg string, keysAndValues ...interface{}) {
	DefaultLogger.Fatalw(msg, keysAndValues...)
}

//
// Debugf
// @Description:
// @param msg
// @param args
//
func Debugf(msg string, args ...interface{}) {
	DefaultLogger.Debugf(msg, args...)
}

//
// Infof
// @Description:
// @param msg
// @param args
//
func Infof(msg string, args ...interface{}) {
	DefaultLogger.Infof(msg, args...)
}

//
// Warnf
// @Description:
// @param msg
// @param args
//
func Warnf(msg string, args ...interface{}) {
	DefaultLogger.Warnf(msg, args...)
}

//
// Errorf
// @Description:
// @param msg
// @param args
//
func Errorf(msg string, args ...interface{}) {
	DefaultLogger.Errorf(msg, args...)
}

//
// Panicf
// @Description:
// @param msg
// @param args
//
func Panicf(msg string, args ...interface{}) {
	DefaultLogger.Panicf(msg, args...)
}

//
// DPanicf
// @Description:
// @param msg
// @param args
//
func DPanicf(msg string, args ...interface{}) {
	DefaultLogger.DPanicf(msg, args...)
}

//
// Fatalf
// @Description:
// @param msg
// @param args
//
func Fatalf(msg string, args ...interface{}) {
	DefaultLogger.Fatalf(msg, args...)
}

//
// Log
// @Description:
// @receiver fn
// @param msg
// @param fields
//
func (fn Func) Log(msg string, fields ...Field) {
	fn(msg, fields...)
}

//
// With
// @Description:
// @param fields
// @return *Logger
//
func With(fields ...Field) *Logger {
	return DefaultLogger.With(fields...)
}
