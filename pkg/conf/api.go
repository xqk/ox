package conf

import (
	"io"

	"github.com/davecgh/go-spew/spew"
)

//  Unmarshaller
type Unmarshaller = func([]byte, interface{}) error

var defaultConfiguration = New()

//
// OnChange
// @Description: 注册change回调函数
// @param fn
//
func OnChange(fn func(*Configuration)) {
	defaultConfiguration.OnChange(fn)
}

func OnLoaded(fn func(*Configuration)) {
	defaultConfiguration.OnLoaded(fn)
}

//
// LoadFromDataSource
// @Description:如果数据源支持动态配置，则从数据源加载配置
// @param ds
// @param unmarshaller
// @return error
//
func LoadFromDataSource(ds DataSource, unmarshaller Unmarshaller) error {
	return defaultConfiguration.LoadFromDataSource(ds, unmarshaller)
}

//
// LoadFromReader
// @Description:使用默认的defaultConfiguration从提供的提供程序加载配置
// @param r
// @param unmarshaller
// @return error
//
func LoadFromReader(r io.Reader, unmarshaller Unmarshaller) error {
	return defaultConfiguration.LoadFromReader(r, unmarshaller)
}

//
// Apply
// @Description:
// @param conf
// @return error
//
func Apply(conf map[string]interface{}) error {
	return defaultConfiguration.apply(conf)
}

//
// Reset
// @Description:重置所有默认设置
//
func Reset() {
	defaultConfiguration = New()
}

//
// Traverse
// @Description:
// @param sep
// @return map[string]interface{}
//
func Traverse(sep string) map[string]interface{} {
	return defaultConfiguration.traverse(sep)
}

//
// Debug
// @Description:
// @param sep
//
func Debug(sep string) {
	spew.Dump("Debug", Traverse(sep))
}

//
// Get
// @Description: 返回一个interface。对于特定的值，可以使用一个Get____方法
// @param key
// @return interface{}
//
func Get(key string) interface{} {
	return defaultConfiguration.Get(key)
}

//
// Exists
// @Description:返回键是否存在
// @param key
// @return bool
//
func Exists(key string) bool {
	return defaultConfiguration.Get(key) != nil
}

//
// Set
// @Description:设置key的配置值
// @param key
// @param val
//
func Set(key string, val interface{}) {
	defaultConfiguration.Set(key, val)
}
