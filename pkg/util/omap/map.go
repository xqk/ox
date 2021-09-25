package omap

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"ox/pkg/util/ocast"
	"github.com/mitchellh/mapstructure"
)

// Unmarshaller
type Unmarshaller = func([]byte, interface{}) error

// KeySpliter
var KeySpliter = "."

//
// FlatMap
// @Description:
//
type FlatMap struct {
	data   map[string]interface{}
	mu     sync.RWMutex
	keyMap sync.Map
}

//
// NewFlatMap
// @Description:
// @return *FlatMap
//
func NewFlatMap() *FlatMap {
	return &FlatMap{
		data: make(map[string]interface{}),
	}
}

//
// Load
// @Description:
// @receiver flat
// @param content
// @param unmarshal
// @return error
//
func (flat *FlatMap) Load(content []byte, unmarshal Unmarshaller) error {
	data := make(map[string]interface{})
	if err := unmarshal(content, &data); err != nil {
		return err
	}
	return flat.apply(data)
}

//
// apply
// @Description: 确认
// @receiver flat
// @param data
// @return error
//
func (flat *FlatMap) apply(data map[string]interface{}) error {
	flat.mu.Lock()
	defer flat.mu.Unlock()

	MergeStringMap(flat.data, data)
	var changes = make(map[string]interface{})
	for k, v := range flat.traverse(KeySpliter) {
		orig, ok := flat.keyMap.Load(k)
		if ok && !reflect.DeepEqual(orig, v) {
			changes[k] = v
		}
		flat.keyMap.Store(k, v)
	}

	return nil
}

//
// Set
// @Description: 设置键的值
// @receiver flat
// @param key
// @param val
// @return error
//
func (flat *FlatMap) Set(key string, val interface{}) error {
	paths := strings.Split(key, KeySpliter)
	lastKey := paths[len(paths)-1]
	m := deepSearch(flat.data, paths[:len(paths)-1])
	m[lastKey] = val
	return flat.apply(m)
}

//
// Get
// @Description: 返回键的值
// @receiver flat
// @param key
// @return interface{}
//
func (flat *FlatMap) Get(key string) interface{} {
	return flat.find(key)
}

//
// GetString
// @Description: 以string形式返回键的值
// @receiver flat
// @param key
// @return string
//
func (flat *FlatMap) GetString(key string) string {
	return ocast.ToString(flat.Get(key))
}

//
// GetBool
// @Description: 以bool形式返回键的值
// @receiver flat
// @param key
// @return bool
//
func (flat *FlatMap) GetBool(key string) bool {
	return ocast.ToBool(flat.Get(key))
}

//
// GetInt
// @Description: 以int形式返回键的值
// @receiver flat
// @param key
// @return int
//
func (flat *FlatMap) GetInt(key string) int {
	return ocast.ToInt(flat.Get(key))
}

//
// GetInt64
// @Description: 以int64形式返回键的值
// @receiver flat
// @param key
// @return int64
//
func (flat *FlatMap) GetInt64(key string) int64 {
	return ocast.ToInt64(flat.Get(key))
}

//
// GetFloat64
// @Description: 以float64形式返回键的值
// @receiver flat
// @param key
// @return float64
//
func (flat *FlatMap) GetFloat64(key string) float64 {
	return ocast.ToFloat64(flat.Get(key))
}

//
// GetTime
// @Description: 以time.Time形式返回键的值
// @receiver flat
// @param key
// @return time.Time
//
func (flat *FlatMap) GetTime(key string) time.Time {
	return ocast.ToTime(flat.Get(key))
}

//
// GetDuration
// @Description: 以time.Duration形式返回键的值
// @receiver flat
// @param key
// @return time.Duration
//
func (flat *FlatMap) GetDuration(key string) time.Duration {
	return ocast.ToDuration(flat.Get(key))
}

//
// GetStringSlice
// @Description: 以[]string形式返回键的值
// @receiver flat
// @param key
// @return []string
//
func (flat *FlatMap) GetStringSlice(key string) []string {
	return ocast.ToStringSlice(flat.Get(key))
}

//
// GetSlice
// @Description: 以[]interface{}形式返回键的值
// @receiver flat
// @param key
// @return []interface{}
//
func (flat *FlatMap) GetSlice(key string) []interface{} {
	return ocast.ToSlice(flat.Get(key))
}

//
// GetStringMap
// @Description: 以map[string]interface{}形式返回键的值
// @receiver flat
// @param key
// @return map[string]interface{}
//
func (flat *FlatMap) GetStringMap(key string) map[string]interface{} {
	return ocast.ToStringMap(flat.Get(key))
}

//
// GetStringMapString
// @Description: 以map[string]string形式返回键的值
// @receiver flat
// @param key
// @return map[string]string
//
func (flat *FlatMap) GetStringMapString(key string) map[string]string {
	return ocast.ToStringMapString(flat.Get(key))
}

//
// GetSliceStringMap
// @Description: 以[]map[string]interface{}形式返回键的值
// @receiver flat
// @param key
// @return []map[string]interface{}
//
func (flat *FlatMap) GetSliceStringMap(key string) []map[string]interface{} {
	return ocast.ToSliceStringMap(flat.Get(key))
}

//
// GetStringMapStringSlice
// @Description:以map[string][]string形式返回键的值
// @receiver flat
// @param key
// @return map[string][]string
//
func (flat *FlatMap) GetStringMapStringSlice(key string) map[string][]string {
	return ocast.ToStringMapStringSlice(flat.Get(key))
}

//
// UnmarshalKey
// @Description:获取单个键并将其解组到Struct中
// @receiver flat
// @param key
// @param rawVal
// @param tagName
// @return error
//
func (flat *FlatMap) UnmarshalKey(key string, rawVal interface{}, tagName string) error {
	config := mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     rawVal,
		TagName:    tagName,
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return err
	}
	if key == "" {
		flat.mu.RLock()
		defer flat.mu.RUnlock()
		return decoder.Decode(flat.data)
	}

	value := flat.Get(key)
	if value == nil {
		return fmt.Errorf("invalid key %s, maybe not exist in config", key)
	}

	return decoder.Decode(value)
}

//
// Reset
// @Description:
// @receiver flat
//
func (flat *FlatMap) Reset() {
	flat.mu.Lock()
	defer flat.mu.Unlock()

	flat.data = make(map[string]interface{})
	// erase map
	flat.keyMap.Range(func(key interface{}, value interface{}) bool {
		flat.keyMap.Delete(key)
		return true
	})
}

//
// find
// @Description:
// @receiver flat
// @param key
// @return interface{}
//
func (flat *FlatMap) find(key string) interface{} {
	dd, ok := flat.keyMap.Load(key)
	if ok {
		return dd
	}

	paths := strings.Split(key, KeySpliter)
	flat.mu.RLock()
	defer flat.mu.RUnlock()
	m := DeepSearchInMap(flat.data, paths[:len(paths)-1]...)
	dd = m[paths[len(paths)-1]]
	flat.keyMap.Store(key, dd)
	return dd
}

//
// lookup
// @Description:
// @param prefix
// @param target
// @param data
// @param sep
//
func lookup(prefix string, target map[string]interface{}, data map[string]interface{}, sep string) {
	for k, v := range target {
		pp := fmt.Sprintf("%s%s%s", prefix, sep, k)
		if prefix == "" {
			pp = fmt.Sprintf("%s", k)
		}
		if dd, err := ocast.ToStringMapE(v); err == nil {
			lookup(pp, dd, data, sep)
		} else {
			data[pp] = v
		}
	}
}

//
// traverse
// @Description:
// @receiver flat
// @param sep
// @return map[string]interface{}
//
func (flat *FlatMap) traverse(sep string) map[string]interface{} {
	data := make(map[string]interface{})
	lookup("", flat.data, data, sep)
	return data
}

//
// deepSearch
// @Description:
// @param m
// @param path
// @return map[string]interface{}
//
func deepSearch(m map[string]interface{}, path []string) map[string]interface{} {
	for _, k := range path {
		m2, ok := m[k]
		if !ok {
			m3 := make(map[string]interface{})
			m[k] = m3
			m = m3
			continue
		}
		m3, ok := m2.(map[string]interface{})
		if !ok {
			m3 = make(map[string]interface{})
			m[k] = m3
		}
		m = m3
	}
	return m
}
