package conf

import (
	"errors"
	"io"
	"net/url"
)

var (
	//ErrConfigAddr 没有配置的错误
	ErrConfigAddr = errors.New("没有配置... ")
	// ErrInvalidDataSource 定义没有注册的错误
	ErrInvalidDataSource = errors.New("无效的数据源，请确保方案已注册")
	datasourceBuilders   = make(map[string]DataSourceCreatorFunc)
	configDecoder        = make(map[string]Unmarshaller)
)

// DataSourceCreatorFunc 表示数据源创建器的函数
type DataSourceCreatorFunc func() DataSource

//
// DataSource
// @Description: 数据源结构体
//
type DataSource interface {
	ReadConfig() ([]byte, error)
	IsConfigChanged() <-chan struct{}
	io.Closer
}

//
// Register
// @Description:向注册中心注册一个数据源创建器的函数
// @param scheme
// @param creator
//
func Register(scheme string, creator DataSourceCreatorFunc) {
	datasourceBuilders[scheme] = creator
}

//
// NewDataSource
// @Description: 新的数据源
// @param configAddr
// @return DataSource
// @return error
//
func NewDataSource(configAddr string) (DataSource, error) {
	if configAddr == "" {
		return nil, ErrConfigAddr
	}
	urlObj, err := url.Parse(configAddr)
	if err != nil {
		return nil, err
	}

	var scheme = urlObj.Scheme
	if scheme == "" {
		scheme = "file"
	}

	creatorFunc, exist := datasourceBuilders[scheme]
	if !exist {
		return nil, ErrInvalidDataSource
	}
	return creatorFunc(), nil
}
