package oecho

import (
	"context"
	"net/http"
	"os"
	"github.com/xqk/ox/pkg/olog"

	"net"

	"github.com/xqk/ox/pkg/constant"
	"github.com/xqk/ox/pkg/registry"
	"github.com/xqk/ox/pkg/server"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Server ...
type Server struct {
	*echo.Echo
	config     *Config
	listener   net.Listener
	registerer registry.Registry
}

func newServer(config *Config) (*Server, error) {
	listener, err := net.Listen("tcp", config.Address())
	if err != nil {
		// config.logger.Panic("new oecho server err", olog.FieldErrKind(ecode.ErrKindListenErr), olog.FieldErr(err))
		return nil, errors.Wrapf(err, "create oecho server failed")
	}
	config.Port = listener.Addr().(*net.TCPAddr).Port
	return &Server{
		Echo:     echo.New(),
		config:   config,
		listener: listener,
	}, nil
}

func (s *Server) Healthz() bool {
	if s.Echo.Listener == nil {
		return false
	}

	conn, err := s.Echo.Listener.Accept()
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

// Server implements server.Server interface.
func (s *Server) Serve() error {
	s.Echo.Logger.SetOutput(os.Stdout)
	s.Echo.Debug = s.config.Debug
	s.Echo.HideBanner = true
	s.Echo.StdLogger = olog.OxLogger.StdLog()
	for _, route := range s.Echo.Routes() {
		s.config.logger.Info("add route", olog.FieldMethod(route.Method), olog.String("path", route.Path))
	}
	s.Echo.Listener = s.listener
	err := s.Echo.Start("")
	if err != http.ErrServerClosed {
		return err
	}

	s.config.logger.Info("close echo", olog.FieldAddr(s.config.Address()))
	return nil
}

// Stop implements server.Server interface
// it will terminate echo server immediately
func (s *Server) Stop() error {
	return s.Echo.Close()
}

// GracefulStop implements server.Server interface
// it will stop echo server gracefully
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.Echo.Shutdown(ctx)
}

// Info returns server info, used by governor and consumer balancer
func (s *Server) Info() *server.ServiceInfo {
	serviceAddr := s.listener.Addr().String()
	if s.config.ServiceAddress != "" {
		serviceAddr = s.config.ServiceAddress
	}

	info := server.ApplyOptions(
		server.WithScheme("http"),
		server.WithAddress(serviceAddr),
		server.WithKind(constant.ServiceProvider),
	)
	// info.Name = info.Name + "." + ModName
	return &info
}
