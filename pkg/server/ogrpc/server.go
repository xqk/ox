package ogrpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"ox/pkg/constant"
	"ox/pkg/server"
)

// Server ...
type Server struct {
	*grpc.Server
	listener net.Listener
	*Config
}

func newServer(config *Config) (*Server, error) {
	var streamInterceptors = append(
		[]grpc.StreamServerInterceptor{defaultStreamServerInterceptor(config.logger, config.SlowQueryThresholdInMilli)},
		config.streamInterceptors...,
	)

	var unaryInterceptors = append(
		[]grpc.UnaryServerInterceptor{defaultUnaryServerInterceptor(config.logger, config.SlowQueryThresholdInMilli)},
		config.unaryInterceptors...,
	)

	config.serverOptions = append(config.serverOptions,
		grpc.StreamInterceptor(StreamInterceptorChain(streamInterceptors...)),
		grpc.UnaryInterceptor(UnaryInterceptorChain(unaryInterceptors...)),
	)

	newServer := grpc.NewServer(config.serverOptions...)
	listener, err := net.Listen(config.Network, config.Address())
	if err != nil {
		// config.logger.Panic("new grpc server err", xlog.FieldErrKind(ecode.ErrKindListenErr), xlog.FieldErr(err))
		return nil, fmt.Errorf("create grpc server failed: %v", err)
	}
	config.Port = listener.Addr().(*net.TCPAddr).Port

	return &Server{
		Server:   newServer,
		listener: listener,
		Config:   config,
	}, nil
}

func (s *Server) Healthz() bool {
	conn, err := s.listener.Accept()
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

// Server implements server.Server interface.
func (s *Server) Serve() error {
	err := s.Server.Serve(s.listener)
	return err
}

// Stop implements server.Server interface
// it will terminate echo server immediately
func (s *Server) Stop() error {
	s.Server.Stop()
	return nil
}

// GracefulStop implements server.Server interface
// it will stop echo server gracefully
func (s *Server) GracefulStop(ctx context.Context) error {
	s.Server.GracefulStop()
	return nil
}

// Info returns server info, used by governor and consumer balancer
func (s *Server) Info() *server.ServiceInfo {
	serviceAddress := s.listener.Addr().String()
	if s.Config.ServiceAddress != "" {
		serviceAddress = s.Config.ServiceAddress
	}

	info := server.ApplyOptions(
		server.WithScheme("grpc"),
		server.WithAddress(serviceAddress),
		server.WithKind(constant.ServiceProvider),
	)
	return &info
}
