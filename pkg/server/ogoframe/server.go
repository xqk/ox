package ogoframe

import (
	"context"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"ox/pkg/constant"
	"ox/pkg/olog"
	"ox/pkg/server"
)

//Server is server core struct
type Server struct {
	*ghttp.Server
	config *Config
}

func newServer(config *Config) *Server {
	s := new(Server)
	serve := g.Server()
	serve.SetAddr(config.Address())

	s.Server = serve
	s.config = config

	return s
}

//Serve ..
func (s *Server) Serve() error {
	routes := s.GetRouterArray()

	for i := 0; i < len(routes); i++ {
		s.config.logger.Info("add route ", olog.FieldMethod(routes[i].Method), olog.String("path", routes[i].Route))
	}
	s.Run()

	return nil
}

//Stop ..
func (s *Server) Stop() error {
	return s.Shutdown()
}

//GracefulStop ..
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.Stop()
}

//Info ..
func (s *Server) Info() *server.ServiceInfo {
	serviceAddr := s.config.Address()
	if s.config.ServiceAddress != "" {
		serviceAddr = s.config.ServiceAddress
	}

	info := server.ApplyOptions(
		server.WithScheme("http"),
		server.WithAddress(serviceAddr),
		server.WithKind(constant.ServiceProvider),
	)
	return &info
}

// Healthz
// TODO(roamerlv):
func (s *Server) Healthz() bool {
	return true
}
