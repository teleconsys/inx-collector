package collector

import (
	"collector/pkg/listener"
	"collector/pkg/poi"
	"collector/pkg/storage"
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/inx-app/nodebridge"
)

type Collector struct {
	*logger.WrappedLogger

	NodeBridge      *nodebridge.NodeBridge
	shutdownHandler *shutdown.ShutdownHandler
	Listener        listener.Listener
	Storage         storage.Storage
	POIHandler      poi.POIHandler
}

func NewCollector(log *logger.Logger, bridge *nodebridge.NodeBridge,
	shutdownHandler *shutdown.ShutdownHandler, storageParameters storage.Parameters, listenerParameters listener.Parameters, poiParameters poi.Parameters) (*Collector, error) {
	collector := &Collector{
		WrappedLogger:   logger.NewWrappedLogger(log),
		NodeBridge:      bridge,
		shutdownHandler: shutdownHandler,
	}

	storage, err := storage.NewStorage(storageParameters, collector.WrappedLogger)
	if err != nil {
		return collector, err
	}
	collector.Storage = storage

	poiHandler := poi.NewPOIHandler(poiParameters)
	collector.POIHandler = poiHandler

	listener, err := listener.NewListener(listenerParameters, storage, poiHandler, collector.WrappedLogger)
	if err != nil {
		return collector, err
	}
	collector.Listener = listener

	return collector, nil
}

func (c *Collector) Run(ctx context.Context) error {

	// manage default storage
	exists, err := c.Storage.CheckCreateBucket(c.Storage.DefaultBucketName, ctx)
	if err != nil {
		c.WrappedLogger.LogErrorf("Can't istantiate storage : %w", err)
		return err
	}
	if exists {
		days, err := c.Storage.GetBucketExpirationDays(c.Storage.DefaultBucketName, ctx)
		if err != nil {
			c.WrappedLogger.LogErrorf("Can't istantiate storage : %w", err)
			return err
		}
		if days != c.Storage.DefaultBucketExpirationDays {
			err = fmt.Errorf("default bucket already exists, but expiration days are %d instead of the specified %d", days, c.Storage.DefaultBucketExpirationDays)
			c.WrappedLogger.LogErrorf("Can't istantiate storage : %w", err)
			return err
		}
	} else {
		err := c.Storage.SetBucketExpirationDays(c.Storage.DefaultBucketName, c.Storage.DefaultBucketExpirationDays, ctx)
		if err != nil {
			c.WrappedLogger.LogErrorf("Can't istantiate storage : %w", err)
			return err
		}
	}

	// load startup filters
	err = c.Listener.LoadStartupFilters(ctx)
	if err != nil {
		c.WrappedLogger.LogErrorf("Can't deploy startup filters : %w", err)
		return err
	}

	// run listener
	client := c.NodeBridge.Client()
	c.WrappedLogger.LogInfo("Running Listener ...")
	err = c.Listener.Run(client, ctx)
	if err != nil {
		c.WrappedLogger.LogErrorf("Running Listener ... exit on error: %w", err)
		return err
	}

	return nil
}
