package collector

import (
	"collector/pkg/api"
	"collector/pkg/collector"
	"collector/pkg/daemon"
	"context"
	"errors"
	"fmt"
	"net/http"

	"time"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/inx-app/httpserver"
	"github.com/iotaledger/inx-app/nodebridge"
	"github.com/labstack/echo/v4"
)

const (
	APIRoute = "collector/v1"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:     "Collector",
			DepsFunc: func(cDeps dependencies) { deps = cDeps },
			Params:   params,
			Provide:  provide,
			Run:      run,
		},
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
	CoreComponent *app.CoreComponent
	deps          dependencies
)

func provide(c *dig.Container) error {

	type inDeps struct {
		dig.In
		NodeBridge *nodebridge.NodeBridge
		*shutdown.ShutdownHandler
	}

	if err := c.Provide(func(deps inDeps) (*collector.Collector, error) {

		return collector.NewCollector(
			CoreComponent.Logger(),
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
			CoreComponent.Logger(),
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
	if err := CoreComponent.Daemon().BackgroundWorker("Collector", func(ctx context.Context) {
		CoreComponent.LogInfo("Starting Collector ...")

		go func() {
			err := deps.Collector.Run(ctx)
			if err != nil {
				deps.ShutdownHandler.SelfShutdown(fmt.Sprintf("Collector shut down, error: %s", err), false)
			}
		}()
		close(collectorInitWait)

	}, daemon.PriorityStopCollector); err != nil {
		CoreComponent.LogPanicf("failed to start worker: %s", err)
	}

	// create a background worker that handles the API
	if err := CoreComponent.Daemon().BackgroundWorker("API", func(ctx context.Context) {
		CoreComponent.LogInfo("Starting API ...")

		// we need to wait until the collector is initialized before starting the API or the daemon is canceled before that is done.
		select {
		case <-ctx.Done():
			return
		case <-collectorInitWait:
		}

		CoreComponent.LogInfo("Starting API ... done")
		CoreComponent.LogInfo("Starting API server ...")

		_ = api.NewServer(deps.Collector, deps.Echo, deps.Collector.WrappedLogger, ctx)

		go func() {
			if err := deps.Echo.Start(ParamsRestAPI.BindAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
				CoreComponent.LogErrorfAndExit("Stopped REST-API server due to an error (%s)", err)
			}
		}()

		ctxRegister, cancelRegister := context.WithTimeout(ctx, 5*time.Second)

		advertisedAddress := ParamsRestAPI.BindAddress
		if ParamsRestAPI.AdvertiseAddress != "" {
			advertisedAddress = ParamsRestAPI.AdvertiseAddress
		}

		if err := deps.NodeBridge.RegisterAPIRoute(ctxRegister, APIRoute, advertisedAddress); err != nil {
			CoreComponent.LogErrorfAndExit("Registering INX api route failed: %s", err)
		}
		cancelRegister()

		CoreComponent.LogInfo("Starting API server ... done")
		<-ctx.Done()
		CoreComponent.LogInfo("Stopping API ...")

		ctxUnregister, cancelUnregister := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelUnregister()

		if err := deps.NodeBridge.UnregisterAPIRoute(ctxUnregister, APIRoute); err != nil {
			CoreComponent.LogWarnf("Unregistering INX api route failed: %s", err)
		}

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCtxCancel()

		if err := deps.Echo.Shutdown(shutdownCtx); err != nil {
			CoreComponent.LogWarn(err)
		}

		CoreComponent.LogInfo("Stopping API ... done")

	}, daemon.PriorityStopRestAPI); err != nil {
		CoreComponent.LogPanicf("failed to start worker: %s", err)
	}

	return nil
}
