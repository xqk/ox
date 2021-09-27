package registry

import (
	"context"
	"github.com/xqk/ox/pkg/olog"
	"github.com/xqk/ox/pkg/server"
)

// Nop registry, used for local development/debugging
type Local struct{}

// ListServices ...
func (n Local) ListServices(ctx context.Context, s string, s2 string) ([]*server.ServiceInfo, error) {
	panic("implement me")
}

// WatchServices ...
func (n Local) WatchServices(ctx context.Context, s string, s2 string) (chan Endpoints, error) {
	panic("implement me")
}

// RegisterService ...
func (n Local) RegisterService(ctx context.Context, si *server.ServiceInfo) error {
	olog.Info("register service locally", olog.FieldMod("registry"), olog.FieldName(si.Name), olog.FieldAddr(si.Label()))
	return nil
}

// UnregisterService ...
func (n Local) UnregisterService(ctx context.Context, si *server.ServiceInfo) error {
	olog.Info("unregister service locally", olog.FieldMod("registry"), olog.FieldName(si.Name), olog.FieldAddr(si.Label()))
	return nil
}

// Close ...
func (n Local) Close() error { return nil }

// Close ...
func (n Local) Kind() string { return "local" }
