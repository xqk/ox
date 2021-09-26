package etcdv3

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"net"
	"net/url"
	"ox/pkg"
	"ox/pkg/client/etcdv3"
	"ox/pkg/constant"
	"ox/pkg/ecode"
	"ox/pkg/olog"
	"ox/pkg/registry"
	"ox/pkg/server"
	"ox/pkg/util/ogo"
	"strings"
	"sync"
	"time"
)

type etcdv3Registry struct {
	client *etcdv3.Client
	kvs    sync.Map
	*Config
	cancel   context.CancelFunc
	rmu      *sync.RWMutex
	sessions map[string]*concurrency.Session
}

func newETCDRegistry(config *Config) (*etcdv3Registry, error) {
	if config.logger == nil {
		config.logger = olog.OxLogger
	}
	config.logger = config.logger.With(olog.FieldMod(ecode.ModRegistryETCD), olog.FieldAddrAny(config.Config.Endpoints))
	etcdv3Client, err := config.Config.Build()
	if err != nil {
		return nil, err
	}
	reg := &etcdv3Registry{
		client:   etcdv3Client,
		Config:   config,
		kvs:      sync.Map{},
		rmu:      &sync.RWMutex{},
		sessions: make(map[string]*concurrency.Session),
	}
	return reg, nil
}

func (reg *etcdv3Registry) Kind() string { return "etcdv3" }

// RegisterService register service to registry
func (reg *etcdv3Registry) RegisterService(ctx context.Context, info *server.ServiceInfo) error {
	err := reg.registerBiz(ctx, info)
	if err != nil {
		return err
	}
	return reg.registerMetric(ctx, info)
}

// UnregisterService unregister service from registry
func (reg *etcdv3Registry) UnregisterService(ctx context.Context, info *server.ServiceInfo) error {
	return reg.unregister(ctx, reg.registerKey(info))
}

// ListServices list service registered in registry with name `name`
func (reg *etcdv3Registry) ListServices(ctx context.Context, name string, scheme string) (services []*server.ServiceInfo, err error) {
	target := fmt.Sprintf("/%s/%s/providers/%s://", reg.Prefix, name, scheme)
	getResp, getErr := reg.client.Get(ctx, target, clientv3.WithPrefix())
	if getErr != nil {
		reg.logger.Error(ecode.MsgWatchRequestErr, olog.FieldErrKind(ecode.ErrKindRequestErr), olog.FieldErr(getErr), olog.FieldAddr(target))
		return nil, getErr
	}

	for _, kv := range getResp.Kvs {
		var service server.ServiceInfo
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			reg.logger.Warnf("invalid service", olog.FieldErr(err))
			continue
		}
		services = append(services, &service)
	}

	return
}

// WatchServices watch service change event, then return address list
func (reg *etcdv3Registry) WatchServices(ctx context.Context, name string, scheme string) (chan registry.Endpoints, error) {
	prefix := fmt.Sprintf("/%s/%s/", reg.Prefix, name)
	watch, err := reg.client.WatchPrefix(context.Background(), prefix)
	if err != nil {
		return nil, err
	}

	var addresses = make(chan registry.Endpoints, 10)
	var al = &registry.Endpoints{
		Nodes:           make(map[string]server.ServiceInfo),
		RouteConfigs:    make(map[string]registry.RouteConfig),
		ConsumerConfigs: make(map[string]registry.ConsumerConfig),
		ProviderConfigs: make(map[string]registry.ProviderConfig),
	}

	for _, kv := range watch.IncipientKeyValues() {
		updateAddrList(al, prefix, scheme, kv)
	}

	// var snapshot registry.Endpoints
	// xstruct.CopyStruct(al, &snapshot)
	addresses <- *al.DeepCopy()

	ogo.Go(func() {
		for event := range watch.C() {
			switch event.Type {
			case mvccpb.PUT:
				updateAddrList(al, prefix, scheme, event.Kv)
			case mvccpb.DELETE:
				deleteAddrList(al, prefix, scheme, event.Kv)
			}

			// var snapshot registry.Endpoints
			// xstruct.CopyStruct(al, &snapshot)
			out := al.DeepCopy()
			select {
			// case addresses <- snapshot:
			case addresses <- *out:
			default:
				olog.Warnf("invalid")
			}
		}
	})

	return addresses, nil
}

func (reg *etcdv3Registry) unregister(ctx context.Context, key string) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, reg.ReadTimeout)
		defer cancel()
	}

	if err := reg.delSession(key); err != nil {
		return err
	}

	_, err := reg.client.Delete(ctx, key)
	if err == nil {
		reg.kvs.Delete(key)
	}
	return err
}

// Close ...
func (reg *etcdv3Registry) Close() error {
	if reg.cancel != nil {
		reg.cancel()
	}
	var wg sync.WaitGroup
	reg.kvs.Range(func(k, v interface{}) bool {
		wg.Add(1)
		go func(k interface{}) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := reg.unregister(ctx, k.(string))
			if err != nil {
				reg.logger.Error("unregister service", olog.FieldErrKind(ecode.ErrKindRequestErr), olog.FieldErr(err), olog.FieldErr(err), olog.FieldKeyAny(k), olog.FieldValueAny(v))
			} else {
				reg.logger.Info("unregister service", olog.FieldKeyAny(k), olog.FieldValueAny(v))
			}
			cancel()
		}(k)
		return true
	})
	wg.Wait()
	return nil
}

func (reg *etcdv3Registry) registerMetric(ctx context.Context, info *server.ServiceInfo) error {
	if info.Kind != constant.ServiceGovernor {
		return nil
	}

	metric := "/prometheus/job/%s/%s/%s"

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, reg.ReadTimeout)
		defer cancel()
	}

	val := info.Address
	key := fmt.Sprintf(metric, info.Name, pkg.HostName(), val)

	opOptions := make([]clientv3.OpOption, 0)
	// opOptions = append(opOptions, clientv3.WithSerializable())
	if ttl := reg.Config.ServiceTTL.Seconds(); ttl > 0 {
		//todo ctx without timeout for same as service life?
		sess, err := reg.getSession(key, concurrency.WithTTL(int(ttl)))
		if err != nil {
			return err
		}
		opOptions = append(opOptions, clientv3.WithLease(sess.Lease()))
	}
	_, err := reg.client.Put(ctx, key, val, opOptions...)
	if err != nil {
		reg.logger.Error("register service", olog.FieldErrKind(ecode.ErrKindRegisterErr), olog.FieldErr(err), olog.FieldKeyAny(key), olog.FieldValueAny(info))
		return err
	}

	reg.logger.Info("register service", olog.FieldKeyAny(key), olog.FieldValueAny(val))
	reg.kvs.Store(key, val)
	return nil

}
func (reg *etcdv3Registry) registerBiz(ctx context.Context, info *server.ServiceInfo) error {
	if _, ok := ctx.Deadline(); !ok {
		var readCancel context.CancelFunc
		ctx, readCancel = context.WithTimeout(ctx, reg.ReadTimeout)
		defer readCancel()
	}

	key := reg.registerKey(info)
	val := reg.registerValue(info)

	opOptions := make([]clientv3.OpOption, 0)
	// opOptions = append(opOptions, clientv3.WithSerializable())
	if ttl := reg.Config.ServiceTTL.Seconds(); ttl > 0 {
		//todo ctx without timeout for same as service life?
		sess, err := reg.getSession(key, concurrency.WithTTL(int(ttl)))
		if err != nil {
			return err
		}
		opOptions = append(opOptions, clientv3.WithLease(sess.Lease()))
	}
	_, err := reg.client.Put(ctx, key, val, opOptions...)
	if err != nil {
		reg.logger.Error("register service", olog.FieldErrKind(ecode.ErrKindRegisterErr), olog.FieldErr(err), olog.FieldKeyAny(key), olog.FieldValueAny(info))
		return err
	}
	reg.logger.Info("register service", olog.FieldKeyAny(key), olog.FieldValueAny(val))
	reg.kvs.Store(key, val)
	return nil
}

func (reg *etcdv3Registry) getSession(k string, opts ...concurrency.SessionOption) (*concurrency.Session, error) {
	reg.rmu.RLock()
	sess, ok := reg.sessions[k]
	reg.rmu.RUnlock()
	if ok {
		return sess, nil
	}
	sess, err := concurrency.NewSession(reg.client.Client)
	if err != nil {
		return sess, err
	}
	reg.rmu.Lock()
	reg.sessions[k] = sess
	reg.rmu.Unlock()
	return sess, nil
}

func (reg *etcdv3Registry) delSession(k string) error {
	if ttl := reg.Config.ServiceTTL.Seconds(); ttl > 0 {
		reg.rmu.RLock()
		sess, ok := reg.sessions[k]
		reg.rmu.RUnlock()
		if ok {
			reg.rmu.Lock()
			delete(reg.sessions, k)
			reg.rmu.Unlock()
			if err := sess.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (reg *etcdv3Registry) registerKey(info *server.ServiceInfo) string {
	return registry.GetServiceKey(reg.Prefix, info)
}

func (reg *etcdv3Registry) registerValue(info *server.ServiceInfo) string {
	return registry.GetServiceValue(info)
}

func deleteAddrList(al *registry.Endpoints, prefix, scheme string, kvs ...*mvccpb.KeyValue) {
	for _, kv := range kvs {
		var addr = strings.TrimPrefix(string(kv.Key), prefix)
		if strings.HasPrefix(addr, "providers/"+scheme) {
			// 解析服务注册键
			addr = strings.TrimPrefix(addr, "providers/")
			if addr == "" {
				continue
			}
			uri, err := url.Parse(addr)
			if err != nil {
				olog.Error("parse uri", olog.FieldErrKind(ecode.ErrKindUriErr), olog.FieldErr(err), olog.FieldKey(string(kv.Key)))
				continue
			}
			delete(al.Nodes, uri.String())
		}

		if strings.HasPrefix(addr, "configurators/"+scheme) {
			// 解析服务配置键
			addr = strings.TrimPrefix(addr, "configurators/")
			if addr == "" {
				continue
			}
			uri, err := url.Parse(addr)
			if err != nil {
				olog.Error("parse uri", olog.FieldErrKind(ecode.ErrKindUriErr), olog.FieldErr(err), olog.FieldKey(string(kv.Key)))
				continue
			}
			delete(al.RouteConfigs, uri.String())
		}

		if isIPPort(addr) {
			// 直接删除addr 因为Delete操作的value值为空
			delete(al.Nodes, addr)
			delete(al.RouteConfigs, addr)
		}
	}
}

func updateAddrList(al *registry.Endpoints, prefix, scheme string, kvs ...*mvccpb.KeyValue) {
	for _, kv := range kvs {
		var addr = strings.TrimPrefix(string(kv.Key), prefix)
		switch {
		// 解析服务注册键
		case strings.HasPrefix(addr, "providers/"+scheme):
			addr = strings.TrimPrefix(addr, "providers/")
			uri, err := url.Parse(addr)
			if err != nil {
				olog.Error("parse uri", olog.FieldErrKind(ecode.ErrKindUriErr), olog.FieldErr(err), olog.FieldKey(string(kv.Key)))
				continue
			}
			var serviceInfo server.ServiceInfo
			if err := json.Unmarshal(kv.Value, &serviceInfo); err != nil {
				olog.Error("parse uri", olog.FieldErrKind(ecode.ErrKindUriErr), olog.FieldErr(err), olog.FieldKey(string(kv.Key)))
				continue
			}
			if serviceInfo.Enable {
				al.Nodes[uri.String()] = serviceInfo
			} else {
				delete(al.Nodes, uri.String())
			}

		case strings.HasPrefix(addr, "configurators/"+scheme):
			addr = strings.TrimPrefix(addr, "configurators/")

			uri, err := url.Parse(addr)
			if err != nil {
				olog.Error("parse uri", olog.FieldErrKind(ecode.ErrKindUriErr), olog.FieldErr(err), olog.FieldKey(string(kv.Key)))
				continue
			}

			if strings.HasPrefix(uri.Path, "/routes/") { // 路由配置
				var routeConfig registry.RouteConfig
				if err := json.Unmarshal(kv.Value, &routeConfig); err != nil {
					olog.Error("parse uri", olog.FieldErrKind(ecode.ErrKindUriErr), olog.FieldErr(err), olog.FieldKey(string(kv.Key)))
					continue
				}
				routeConfig.ID = strings.TrimPrefix(uri.Path, "/routes/")
				routeConfig.Scheme = uri.Scheme
				routeConfig.Host = uri.Host
				al.RouteConfigs[uri.String()] = routeConfig
			}

			if strings.HasPrefix(uri.Path, "/providers/") {
				var providerConfig registry.ProviderConfig
				if err := json.Unmarshal(kv.Value, &providerConfig); err != nil {
					olog.Error("parse uri", olog.FieldErrKind(ecode.ErrKindUriErr), olog.FieldErr(err), olog.FieldKey(string(kv.Key)))
					continue
				}
				providerConfig.ID = strings.TrimPrefix(uri.Path, "/providers/")
				providerConfig.Scheme = uri.Scheme
				providerConfig.Host = uri.Host
				al.ProviderConfigs[uri.String()] = providerConfig
			}

			if strings.HasPrefix(uri.Path, "/consumers/") {
				var consumerConfig registry.ConsumerConfig
				if err := json.Unmarshal(kv.Value, &consumerConfig); err != nil {
					olog.Error("parse uri", olog.FieldErrKind(ecode.ErrKindUriErr), olog.FieldErr(err), olog.FieldKey(string(kv.Key)))
					continue
				}
				consumerConfig.ID = strings.TrimPrefix(uri.Path, "/consumers/")
				consumerConfig.Scheme = uri.Scheme
				consumerConfig.Host = uri.Host
				al.ConsumerConfigs[uri.String()] = consumerConfig
			}
		}
	}
}

func isIPPort(addr string) bool {
	_, _, err := net.SplitHostPort(addr)
	return err == nil
}

/*
key: /ox/main/configurator/grpc:///routes/1
val:
{
	"upstream": { // 客户端配置
		"nodes": { // 按照node负载均衡
			"127.0.0.1:1980": 1,
			"127.0.0.1:1981": 4
		},
		"group": { // 按照group负载均衡
			"red": 2,
			"green": 1
		}
	},
	"uri": "/hello",
	"deployment": "open_api"
}

key: /ox/main/configurator/grpc://127.0.0.1/routes/2
val:
{
	"upstream": { // 客户端配置
		"nodes": { // 按照node负载均衡
			"127.0.0.1:1980": 1,
			"127.0.0.1:1981": 1
		},
		"group": { // 按照group负载均衡
			"red": 1,
			"green": 2
		}
	},
	"uri": "/hello",
	"deployment": "core_api" // 部署组
}

key: /ox/main/configurator/grpc:///consumers/client-demo
val:
{

}
*/
