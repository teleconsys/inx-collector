package api

import (
	"collector/pkg/collector"
	"context"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/labstack/echo/v4"
)

const (
	APIRoute = "/api/collector/v1"
)

type Server struct {
	*logger.WrappedLogger
	Collector *collector.Collector
	Context   context.Context
}

func NewServer(collector *collector.Collector, echo *echo.Echo, log *logger.WrappedLogger, ctx context.Context) *Server {
	s := &Server{
		WrappedLogger: logger.NewWrappedLogger(log.LoggerNamed("ServerRestAPI")),
		Collector:     collector,
		Context:       ctx,
	}
	s.setupRoutes(echo.Group(APIRoute))
	return s
}
