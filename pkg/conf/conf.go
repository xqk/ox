package conf

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"ox/pkg/util/ocast"
	"ox/pkg/util/omap"
)

//
// Configuration
// @Description:为应用程序提供配置
//
type Configuration struct {
	mu       sync.RWMutex
	override map[string]interface{}
	keyDelim string

	keyMap    *sync.Map
	onChanges []func(*Configuration)
	onLoadeds []func(*Configuration)

	watchers map[string][]func(*Configuration)
	// TODO: concurrency protect
	loaded bool
}

const (
	defaultKeyDelim = "."
)

//
// New
// @Description:使用程序构造一个新的配置
// @return *Configuration
//
func New() *Configuration {
	return &Configuration{
		override:  make(map[string]interface{}),
		keyDelim:  defaultKeyDelim,
		keyMap:    &sync.Map{},
		onChanges: make([]func(*Configuration), 0),
		onLoadeds: make([]func(*Configuration), 0),
		watchers:  make(map[string][]func(*Configuration)),
		loaded:    false,
	}
}

//
// SetKeyDelim
// @Description:设置defaultConfiguration实例的keyDelim值
// @receiver c
// @param delim
//
func (c *Configuration) SetKeyDelim(delim string) {
	c.keyDelim = delim
}

//
// Sub
// @Description:返回表示此实例的子树的新Configuration实例
// @receiver c
// @param key
// @return *Configuration
//
func (c *Configuration) Sub(key string) *Configuration {
	return &Configuration{
		keyDelim: c.keyDelim,
		override: c.GetStringMap(key),
	}
}

//
// WriteConfig
// @Description: 写入配置
// @receiver c
// @return error
//
func (c *Configuration) WriteConfig() error {
	// return c.provider.Write(c.override)
	return nil
}

//
// OnChange
// @Description: 注册change回调函数
// @receiver c
// @param fn
//
func (c *Configuration) OnChange(fn func(*Configuration)) {
	c.onChanges = append(c.onChanges, fn)
}

//
// OnLoaded
// @Description:
// @receiver c
// @param fn
//
func (c *Configuration) OnLoaded(fn func(*Configuration)) {
	if c.loaded {
		fn(c)
		return
	}

	c.onLoadeds = append(c.onLoadeds, fn)
}

//
// LoadEnvironments
// @Description:带有前缀(如APP_ PREFIX_FIELD1_FIELD2)的操作系统的环境变量将被转换为prefix.field1.field2
// @receiver c
// @param prefix
//
func (c *Configuration) LoadEnvironments(prefix string) {
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}
		key := strings.ToLower(strings.ReplaceAll(env, "_", "."))
		val := os.Getenv(env)
		c.Set(key, val)
	}
}

//
// LoadFromDataSource
// @Description:从数据源加载数据
// @receiver c
// @param ds
// @param unmarshaller
// @return error
//
func (c *Configuration) LoadFromDataSource(ds DataSource, unmarshaller Unmarshaller) error {
	content, err := ds.ReadConfig()
	if err != nil {
		return err
	}

	if err := c.Load(content, unmarshaller); err != nil {
		return err
	}

	go func() {
		for range ds.IsConfigChanged() {
			if content, err := ds.ReadConfig(); err == nil {
				_ = c.reflush(content, unmarshaller)
				for _, change := range c.onChanges {
					change(c)
				}
			}
		}
	}()

	return nil
}

//
// reflush
// @Description:重新刷新
// @receiver c
// @param content
// @param unmarshal
// @return error
//
func (c *Configuration) reflush(content []byte, unmarshal Unmarshaller) error {
	configuration := make(map[string]interface{})
	if err := unmarshal(content, &configuration); err != nil {
		return err
	}
	if err := c.apply(configuration); err != nil {
		return err
	}

	return nil
}

//
// Load
// @Description: 加载
// @receiver c
// @param content
// @param unmarshal
// @return error
//
func (c *Configuration) Load(content []byte, unmarshal Unmarshaller) error {
	if err := c.reflush(content, unmarshal); err != nil {
		return err
	}

	log.Print("load config successfully")
	c.loaded = true
	for _, loadHook := range c.onLoadeds {
		loadHook(c)
	}
	return nil
}

//
// LoadFromReader
// @Description:从提供的数据源加载配置
// @receiver c
// @param reader
// @param unmarshaller
// @return error
//
func (c *Configuration) LoadFromReader(reader io.Reader, unmarshaller Unmarshaller) error {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return c.Load(content, unmarshaller)
}

//
// apply
// @Description:应用配置
// @receiver c
// @param conf
// @return error
//
func (c *Configuration) apply(conf map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var changes = make(map[string]interface{})

	omap.MergeStringMap(c.override, conf)
	for k, v := range c.traverse(c.keyDelim) {
		orig, ok := c.keyMap.Load(k)
		if ok && !reflect.DeepEqual(orig, v) {
			changes[k] = v
		}
		c.keyMap.Store(k, v)
	}

	if len(changes) > 0 {
		c.notifyChanges(changes)
	}

	return nil
}

//
// notifyChanges
// @Description:通知更改
// @receiver c
// @param changes
//
func (c *Configuration) notifyChanges(changes map[string]interface{}) {
	var changedWatchPrefixMap = map[string]struct{}{}

	for watchPrefix := range c.watchers {
		for key := range changes {
			// 前缀匹配即可
			// todo 可能产生错误匹配
			if strings.HasPrefix(key, watchPrefix) {
				changedWatchPrefixMap[watchPrefix] = struct{}{}
			}
		}
	}

	for changedWatchPrefix := range changedWatchPrefixMap {
		for _, handle := range c.watchers[changedWatchPrefix] {
			go handle(c)
		}
	}
}

//
// Set
// @Description: 设置配置
// @receiver c
// @param key
// @param val
// @return error
//
func (c *Configuration) Set(key string, val interface{}) error {
	paths := strings.Split(key, c.keyDelim)
	lastKey := paths[len(paths)-1]
	m := deepSearch(c.override, paths[:len(paths)-1])
	m[lastKey] = val
	return c.apply(m)
	// c.keyMap.Store(key, val)
}

//
// deepSearch
// @Description: 深层搜索
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

//
// Get
// @Description:返回与键关联的值
// @receiver c
// @param key
// @return interface{}
//
func (c *Configuration) Get(key string) interface{} {
	return c.find(key)
}

//
// GetString
// @Description:以带有默认defaultConfiguration的字符串形式返回与键关联的值
// @param key
// @return string
//
func GetString(key string) string {
	return defaultConfiguration.GetString(key)
}

//
// GetString
// @Description:以字符串形式返回与键关联的值
// @receiver c
// @param key
// @return string
//
func (c *Configuration) GetString(key string) string {
	return ocast.ToString(c.Get(key))
}

//
//
// GetBool
// @Description:以带有默认defaultConfiguration的bool形式返回与键关联的值
// @param key
// @return bool
//
func GetBool(key string) bool {
	return defaultConfiguration.GetBool(key)
}

//
// GetBool
// @Description: 以布尔值的形式返回与键关联的值
// @receiver c
// @param key
// @return bool
//
func (c *Configuration) GetBool(key string) bool {
	return ocast.ToBool(c.Get(key))
}

//
// GetInt
// @Description:以Int的形式返回与键关联的值
// @param key
// @return int
//
func GetInt(key string) int {
	return defaultConfiguration.GetInt(key)
}

//
// GetInt
// @Description: 以Int的形式返回与键关联的值
// @receiver c
// @param key
// @return int
//
func (c *Configuration) GetInt(key string) int {
	return ocast.ToInt(c.Get(key))
}

//
// GetInt64
// @Description: 以Int64的形式返回与键关联的值
// @param key
// @return int64
//
func GetInt64(key string) int64 {
	return defaultConfiguration.GetInt64(key)
}

//
// GetInt64
// @Description: 以Int的形式返回与键关联的值
// @receiver c
// @param key
// @return int64
//
func (c *Configuration) GetInt64(key string) int64 {
	return ocast.ToInt64(c.Get(key))
}

//
// GetFloat64
// @Description: 以Float64的形式返回与键关联的值
// @param key
// @return float64
//
func GetFloat64(key string) float64 {
	return defaultConfiguration.GetFloat64(key)
}

//
// GetFloat64
// @Description: 以Float64的形式返回与键关联的值
// @receiver c
// @param key
// @return float64
//
func (c *Configuration) GetFloat64(key string) float64 {
	return ocast.ToFloat64(c.Get(key))
}

//
// GetTime
// @Description: 以Time的形式返回与键关联的值
// @param key
// @return time.Time
//
func GetTime(key string) time.Time {
	return defaultConfiguration.GetTime(key)
}

//
// GetTime
// @Description: 以Time的形式返回与键关联的值
// @receiver c
// @param key
// @return time.Time
//
func (c *Configuration) GetTime(key string) time.Time {
	return ocast.ToTime(c.Get(key))
}

//
// GetDuration
// @Description: 以Duration的形式返回与键关联的值
// @param key
// @return time.Duration
//
func GetDuration(key string) time.Duration {
	return defaultConfiguration.GetDuration(key)
}

//
// GetDuration
// @Description: 以Duration的形式返回与键关联的值
// @receiver c
// @param key
// @return time.Duration
//
func (c *Configuration) GetDuration(key string) time.Duration {
	return ocast.ToDuration(c.Get(key))
}

//
// GetStringSlice
// @Description: 以String数组的形式返回与键关联的值
// @param key
// @return []string
//
func GetStringSlice(key string) []string {
	return defaultConfiguration.GetStringSlice(key)
}

//
// GetStringSlice
// @Description: 以String数组的形式返回与键关联的值
// @receiver c
// @param key
// @return []string
//
func (c *Configuration) GetStringSlice(key string) []string {
	return ocast.ToStringSlice(c.Get(key))
}

//
// GetSlice
// @Description: 以interface{}数组的形式返回与键关联的值
// @param key
// @return []interface{}
//
func GetSlice(key string) []interface{} {
	return defaultConfiguration.GetSlice(key)
}

//
// GetSlice
// @Description: 以interface{}数组的形式返回与键关联的值
// @receiver c
// @param key
// @return []interface{}
//
func (c *Configuration) GetSlice(key string) []interface{} {
	return ocast.ToSlice(c.Get(key))
}

//
// GetStringMap
// @Description: 以StringMap的形式返回与键关联的值
// @param key
// @return map[string]interface{}
//
func GetStringMap(key string) map[string]interface{} {
	return defaultConfiguration.GetStringMap(key)
}

//
// GetStringMap
// @Description: 以StringMap的形式返回与键关联的值
// @receiver c
// @param key
// @return map[string]interface{}
//
func (c *Configuration) GetStringMap(key string) map[string]interface{} {
	return ocast.ToStringMap(c.Get(key))
}

//
// GetStringMapString
// @Description: 以map[string]string的形式返回与键关联的值
// @param key
// @return map[string]string
//
func GetStringMapString(key string) map[string]string {
	return defaultConfiguration.GetStringMapString(key)
}

//
// GetStringMapString
// @Description: 以map[string]string的形式返回与键关联的值
// @receiver c
// @param key
// @return map[string]string
//
func (c *Configuration) GetStringMapString(key string) map[string]string {
	return ocast.ToStringMapString(c.Get(key))
}

//
// GetSliceStringMap
// @Description: 以[]map[string]interface{}的形式返回与键关联的值
// @receiver c
// @param key
// @return []map[string]interface{}
//
func (c *Configuration) GetSliceStringMap(key string) []map[string]interface{} {
	return ocast.ToSliceStringMap(c.Get(key))
}

//
// GetStringMapStringSlice
// @Description: 以map[string][]string的形式返回与键关联的值
// @param key
// @return map[string][]string
//
func GetStringMapStringSlice(key string) map[string][]string {
	return defaultConfiguration.GetStringMapStringSlice(key)
}

//
// GetStringMapStringSlice
// @Description: 以map[string][]string的形式返回与键关联的值
// @receiver c
// @param key
// @return map[string][]string
//
func (c *Configuration) GetStringMapStringSlice(key string) map[string][]string {
	return ocast.ToStringMapStringSlice(c.Get(key))
}

//
// UnmarshalWithExpect
// @Description: Unmarshal键，如果失败返回expect
// @param key
// @param expect
// @return interface{}
//
func UnmarshalWithExpect(key string, expect interface{}) interface{} {
	return defaultConfiguration.UnmarshalWithExpect(key, expect)
}

//
// UnmarshalWithExpect
// @Description: Unmarshal键，如果失败返回expect
// @receiver c
// @param key
// @param expect
// @return interface{}
//
func (c *Configuration) UnmarshalWithExpect(key string, expect interface{}) interface{} {
	err := c.UnmarshalKey(key, expect)
	if err != nil {
		return expect
	}
	return expect
}

//
// UnmarshalKey
// @Description: 使用默认的defaultConfiguration将单个键解封到Struct中
// @param key
// @param rawVal
// @param opts
// @return error
//
func UnmarshalKey(key string, rawVal interface{}, opts ...GetOption) error {
	return defaultConfiguration.UnmarshalKey(key, rawVal, opts...)
}

// ErrInvalidKey 无效键的错误
var ErrInvalidKey = errors.New("无效的key，可能在配置中不存在")

//
// UnmarshalKey
// @Description:获取单个键并将其解组到Struct中
// @receiver c
// @param key
// @param rawVal
// @param opts
// @return error
//
func (c *Configuration) UnmarshalKey(key string, rawVal interface{}, opts ...GetOption) error {
	var options = defaultGetOptions
	for _, opt := range opts {
		opt(&options)
	}

	config := mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     rawVal,
		TagName:    options.TagName,
	}
	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return err
	}
	if key == "" {
		c.mu.RLock()
		defer c.mu.RUnlock()
		return decoder.Decode(c.override)
	}

	value := c.Get(key)
	if value == nil {
		return errors.Wrap(ErrInvalidKey, key)
	}

	return decoder.Decode(value)
}

func (c *Configuration) find(key string) interface{} {
	dd, ok := c.keyMap.Load(key)
	if ok {
		return dd
	}

	paths := strings.Split(key, c.keyDelim)
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := omap.DeepSearchInMap(c.override, paths[:len(paths)-1]...)
	dd = m[paths[len(paths)-1]]
	c.keyMap.Store(key, dd)
	return dd
}

func lookup(prefix string, target map[string]interface{}, data map[string]interface{}, sep string) {
	for k, v := range target {
		pp := fmt.Sprintf("%s%s%s", prefix, sep, k)
		if prefix == "" {
			pp = k
		}
		if dd, err := ocast.ToStringMapE(v); err == nil {
			lookup(pp, dd, data, sep)
		} else {
			data[pp] = v
		}
	}
}

func (c *Configuration) traverse(sep string) map[string]interface{} {
	data := make(map[string]interface{})
	lookup("", c.override, data, sep)
	return data
}
