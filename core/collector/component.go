package collector

import (
	"collector/pkg/api"
	"collector/pkg/collector"
	"collector/pkg/daemon"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"time"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/app/shutdown"
	"github.com/iotaledger/inx-app/pkg/httpserver"
	"github.com/iotaledger/inx-app/pkg/nodebridge"
	"github.com/labstack/echo/v4"
)

func init() {
	Component = &app.Component{
		Name:     "Collector",
		DepsFunc: func(cDeps dependencies) { deps = cDeps },
		Params:   params,
		Provide:  provide,
		Run:      run,
	}
}

type dependencies struct {
	dig.In
	NodeBridge      *nodebridge.NodeBridge
	Collector       *collector.Collector
	ShutdownHandler *shutdown.ShutdownHandler
	Echo            *echo.Echo
}

var (
	Component *app.Component
	deps      dependencies
)

func provide(c *dig.Container) error {

	type inDeps struct {
		dig.In
		NodeBridge *nodebridge.NodeBridge
		*shutdown.ShutdownHandler
	}

	if err := c.Provide(func(deps inDeps) (*collector.Collector, error) {

		return collector.NewCollector(
			Component.Logger(),
			deps.NodeBridge,
			deps.ShutdownHandler,
			*ParamsStorage,
			*ParamsListener,
			*ParamsPOI,
		)
	}); err != nil {
		return err
	}

	if err := c.Provide(func() *echo.Echo {
		return httpserver.NewEcho(
			Component.Logger(),
			nil,
			ParamsRestAPI.DebugRequestLoggerEnabled,
		)
	}); err != nil {
		return err
	}

	return nil
}

func run() error {

	collectorInitWait := make(chan struct{})

	// create a background worker that handles the collector events
	if err := Component.Daemon().BackgroundWorker("Collector", func(ctx context.Context) {
		Component.LogInfo("Starting Collector ...")

		go func() {
			err := deps.Collector.Run(ctx)
			if err != nil {
				deps.ShutdownHandler.SelfShutdown(fmt.Sprintf("Collector shut down, error: %s", err), false)
			}
		}()
		close(collectorInitWait)

	}, daemon.PriorityStopCollector); err != nil {
		Component.LogPanicf("failed to start worker: %s", err)
	}

	// create a background worker that handles the API
	if err := Component.Daemon().BackgroundWorker("API", func(ctx context.Context) {
		Component.LogInfo("Starting API ...")

		// we need to wait until the collector is initialized before starting the API or the daemon is canceled before that is done.
		select {
		case <-ctx.Done():
			return
		case <-collectorInitWait:
		}

		Component.LogInfo("Starting API ... done")
		Component.LogInfo("Starting API server ...")

		_ = api.NewServer(deps.Collector, deps.Echo, deps.Collector.WrappedLogger, ctx)

		go func() {
			if err := deps.Echo.Start(ParamsRestAPI.BindAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
				Component.LogErrorfAndExit("Stopped REST-API server due to an error (%s)", err)
			}
		}()

		ctxRegister, cancelRegister := context.WithTimeout(ctx, 5*time.Second)

		advertisedAddress := ParamsRestAPI.BindAddress
		if ParamsRestAPI.AdvertiseAddress != "" {
			advertisedAddress = ParamsRestAPI.AdvertiseAddress
		}

		routeName := strings.Replace(api.APIRoute, "/api/", "", 1)

		if err := deps.NodeBridge.RegisterAPIRoute(ctxRegister, routeName, advertisedAddress, api.APIRoute); err != nil {
			Component.LogErrorfAndExit("Registering INX api route failed: %s", err)
		}
		cancelRegister()

		Component.LogInfo("Starting API server ... done")
		<-ctx.Done()
		Component.LogInfo("Stopping API ...")

		ctxUnregister, cancelUnregister := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelUnregister()

		if err := deps.NodeBridge.UnregisterAPIRoute(ctxUnregister, routeName); err != nil {
			Component.LogWarnf("Unregistering INX api route failed: %s", err)
		}

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCtxCancel()

		if err := deps.Echo.Shutdown(shutdownCtx); err != nil {
			Component.LogWarn(err)
		}

		Component.LogInfo("Stopping API ... done")

	}, daemon.PriorityStopRestAPI); err != nil {
		Component.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
