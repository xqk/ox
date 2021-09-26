package ogrpc

import (
	"context"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"net"
	"ox/pkg/constant"
	"ox/pkg/olog"
	"testing"
	"time"
)

func TestServer_Serve(t *testing.T) {
	type fields struct {
		Server   *grpc.Server
		listener net.Listener
		Config   *Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Server:   tt.fields.Server,
				listener: tt.fields.listener,
				Config:   tt.fields.Config,
			}
			if err := s.Serve(); (err != nil) != tt.wantErr {
				t.Errorf("Server.Serve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_Closed(t *testing.T) {
	convey.Convey("test server stop", t, func() {
		config := DefaultConfig()
		config.Port = 0
		ns, err := newServer(config)
		assert.Nil(t, err)
		err = ns.Stop()
		convey.So(err, convey.ShouldBeNil)
		err = ns.Serve()
		convey.So(err, convey.ShouldEqual, grpc.ErrServerStopped)
		// server.Serve is responsible for closing the listener, even if the
		// server was already stopped.
		err = ns.listener.Close()
		convey.So(errorDesc(err), convey.ShouldContainSubstring, "use of closed")
	})
}
func TestServer_Stop(t *testing.T) {
	convey.Convey("test server graceful stop", t, func(c convey.C) {
		ns, err := newServer(&Config{
			Network:                   "tcp4",
			Host:                      "127.0.0.1",
			Port:                      0,
			Deployment:                constant.DefaultDeployment,
			DisableMetric:             false,
			DisableTrace:              false,
			SlowQueryThresholdInMilli: 500,
			logger:                    olog.OxLogger.With(olog.FieldMod("server.grpc")),
			serverOptions:             []grpc.ServerOption{},
			streamInterceptors:        []grpc.StreamServerInterceptor{},
			unaryInterceptors:         []grpc.UnaryServerInterceptor{},
		})
		assert.Nil(t, err)
		//
		go func() {
			// make sure Serve() is called
			time.Sleep(time.Millisecond * 500)
			err := ns.Stop()
			c.So(err, convey.ShouldBeNil)
		}()

		err = ns.Serve()
		convey.So(err, convey.ShouldBeNil)
		// server.Serve is responsible for closing the listener, even if the
		// server was already stopped.
		err = ns.listener.Close()
		convey.So(errorDesc(err), convey.ShouldContainSubstring, "use of closed")
	})
}
func TestServer_GracefulStop(t *testing.T) {
	convey.Convey("test server graceful stop", t, func(c convey.C) {
		ns, err := newServer(&Config{
			Network:                   "tcp4",
			Host:                      "127.0.0.1",
			Port:                      0,
			Deployment:                constant.DefaultDeployment,
			DisableMetric:             false,
			DisableTrace:              false,
			SlowQueryThresholdInMilli: 500,
			logger:                    olog.OxLogger.With(olog.FieldMod("server.grpc")),
			serverOptions:             []grpc.ServerOption{},
			streamInterceptors:        []grpc.StreamServerInterceptor{},
			unaryInterceptors:         []grpc.UnaryServerInterceptor{},
		})
		assert.Nil(t, err)
		//
		go func() {
			// make sure Serve() is called
			time.Sleep(time.Millisecond * 500)
			err := ns.GracefulStop(context.TODO())
			c.So(err, convey.ShouldBeNil)
		}()

		err = ns.Serve()
		convey.So(err, convey.ShouldBeNil)
		// server.Serve is responsible for closing the listener, even if the
		// server was already stopped.
		err = ns.listener.Close()
		convey.So(errorDesc(err), convey.ShouldContainSubstring, "use of closed")
	})
}

func TestServer_Info(t *testing.T) {
	convey.Convey("test server info", t, func(c convey.C) {
		ns, err := newServer(&Config{
			Network:                   "tcp4",
			Host:                      "127.0.0.1",
			Port:                      0,
			Deployment:                constant.DefaultDeployment,
			DisableMetric:             false,
			DisableTrace:              false,
			SlowQueryThresholdInMilli: 500,
			logger:                    olog.OxLogger.With(olog.FieldMod("server.grpc")),
			serverOptions:             []grpc.ServerOption{},
			streamInterceptors:        []grpc.StreamServerInterceptor{},
			unaryInterceptors:         []grpc.UnaryServerInterceptor{},
		})
		assert.Nil(t, err)
		convey.So(ns.Info().Scheme, convey.ShouldEqual, "grpc")
		convey.So(ns.Info().Enable, convey.ShouldEqual, true)
	})
}

func errorDesc(err error) string {
	if s, ok := status.FromError(err); ok {
		return s.Message()
	}
	return err.Error()
}
