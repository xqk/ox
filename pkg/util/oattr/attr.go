package oattr

import "errors"

//
// Attributes
// @Description:
//
type Attributes struct {
	m map[interface{}]interface{}
}

var (
	// 无效的key
	ErrInvalidKVPairs = errors.New("invalid kv pairs")
)

//
// New
// @Description:
// @param kvs
// @return *Attributes
//
func New(kvs ...interface{}) *Attributes {
	if len(kvs)%2 != 0 {
		panic(ErrInvalidKVPairs)
	}
	a := &Attributes{m: make(map[interface{}]interface{}, len(kvs)/2)}
	for i := 0; i < len(kvs)/2; i++ {
		a.m[kvs[i*2]] = kvs[i*2+1]
	}
	return a
}

//
// WithValues
// @Description:
// @receiver a
// @param kvs
// @return *Attributes
//
func (a *Attributes) WithValues(kvs ...interface{}) *Attributes {
	if len(kvs)%2 != 0 {
		panic(ErrInvalidKVPairs)
	}
	n := &Attributes{m: make(map[interface{}]interface{}, len(a.m)+len(kvs)/2)}
	for k, v := range a.m {
		n.m[k] = v
	}
	for i := 0; i < len(kvs)/2; i++ {
		n.m[kvs[i*2]] = kvs[i*2+1]
	}
	return n
}

//
// Value
// @Description:
// @receiver a
// @param key
// @return interface{}
//
func (a *Attributes) Value(key interface{}) interface{} {
	return a.m[key]
}