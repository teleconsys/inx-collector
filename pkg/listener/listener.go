package listener

import (
	"collector/pkg/poi"
	"collector/pkg/storage"
	"context"
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/iotaledger/datapayloads.go"
	"github.com/iotaledger/hive.go/core/logger"
	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
)

type Listener struct {
	*logger.WrappedLogger
	Filters        map[string]Filter
	Storage        storage.Storage
	POIHandler     poi.POIHandler
	StartupFilters []Filter
}

func NewListener(params Parameters, storage storage.Storage, poiHandler poi.POIHandler, log *logger.WrappedLogger) (Listener, error) {
	var filters []Filter
	var err error

	if params.Filters != "" {
		filters, err = UnmarshalStartupFilters(params.Filters)
	}

	listener := Listener{
		WrappedLogger:  logger.NewWrappedLogger(log.LoggerNamed("Listener")),
		Filters:        make(map[string]Filter),
		Storage:        storage,
		POIHandler:     poiHandler,
		StartupFilters: filters,
	}
	return listener, err
}

func (l *Listener) Run(client inx.INXClient, ctx context.Context) error {
	// Listen to all referenced blocks
	stream, err := client.ListenToReferencedBlocks(ctx, &inx.NoParams{})
	if err != nil {
		return err
	}

	for {
		newBlock, err := stream.Recv()
		if err != nil {
			l.WrappedLogger.LogErrorf("could not receive block, error: %w", err)
			continue
		}
		// we do something only if we have filters
		if len(l.Filters) == 0 {
			continue
		}
		// get tagged data
		blockId := newBlock.GetBlockId()
		taggedData, block, err := GetTaggedDataFromId(blockId, client, ctx)
		if err != nil {
			l.WrappedLogger.LogErrorf("could not process block, error: %w", err)
			continue
		}
		// starts a routine to manage the tagged payload and keeps listening
		go func(filters map[string]Filter, taggedData iotago.TaggedData, block iotago.Block, blockId inx.BlockId, c context.Context) {
			for filterId := range filters {
				err := l.checkAndStore(taggedData, filterId, &block, blockId, ctx)
				if err != nil {
					l.WrappedLogger.LogErrorf("tagged data error: %w", err)
					continue
				}
			}
		}(l.Filters, taggedData, *block, *blockId, ctx)
	}
}

func (l *Listener) AddFilter(filter Filter) (string, error) {
	if filter.Duration != "" {
		err := filter.setExpiration()
		if err != nil {
			return "", err
		}
	}
	filter.setId()
	for _, f := range l.Filters {
		if f.Id == filter.Id {
			err := fmt.Errorf("filter id '%s' already exists", filter.Id)
			return "", err
		}
	}
	l.Filters[filter.Id] = filter
	l.WrappedLogger.LogInfof("Filter '%s' added, listening on tag: '%s'", filter.Id, filter.Tag)
	return filter.Id, nil
}

func (l *Listener) RemoveFilter(filterId string) error {
	tag := l.Filters[filterId].Tag
	delete(l.Filters, filterId)
	l.WrappedLogger.LogInfof("Filter '%s' added, is no longer listening on tag: '%s'", filterId, tag)
	return nil
}

func (l *Listener) LoadStartupFilters(ctx context.Context) error {
	for _, filter := range l.StartupFilters {
		// use default bucket if none
		if filter.BucketName == "" {
			filter.BucketName = l.Storage.DefaultBucketName
		} else {
			// check if provided bucket exists
			exists, err := l.Storage.BucketExists(filter.BucketName, ctx)
			if err != nil && exists {
				l.WrappedLogger.LogErrorf("Can't deploy startup filters : %w", err)
				return err
			}
			if !exists {
				err = fmt.Errorf("bucket '%s' doesn't exist", filter.BucketName)
				l.WrappedLogger.LogErrorf("Can't deploy startup filters : %w", err)
				return err
			}
		}
		l.AddFilter(filter)
	}
	return nil
}

func (l *Listener) checkFilterExpired(filterId string) bool {
	filter := l.Filters[filterId]
	filterExpired := filter.IsExpired()
	if filterExpired {
		l.RemoveFilter(filterId)
	}
	return filterExpired
}

func (l *Listener) checkAndStore(taggedData iotago.TaggedData, filterId string, block *iotago.Block, blockId inx.BlockId, ctx context.Context) error {
	var err error
	filter := l.Filters[filterId]
	if string(taggedData.Tag) == filter.Tag {
		if filter.Duration != "" {
			// checks if the filter expired, if it is, skips and removes the filter
			if l.checkFilterExpired(filterId) {
				l.WrappedLogger.LogInfof("Filter '%s' expired, with tag: '%s'", filter.Id, filter.Tag)
				return nil
			}
		}

		// checks if the filter has a specified public key, if it does it verifies the data
		if filter.PublicKeyBytes != nil {

			// check if this payload is a signed payload compliant to the filter specification
			signedPayload, err := getSubscribedSignedPayload(taggedData, filter.PublicKeyBytes)
			if err != nil {
				err = fmt.Errorf("not a subscribed payload, %v", err)
				return err
			}

			// verifies signature
			err = signedPayload.VerifySignature()
			if err != nil {
				err = fmt.Errorf("subscribed payload with invalid signature, %v", err)
				return err
			}
		}

		blockIdStr := hex.EncodeToString(blockId.GetId())
		var object storage.Object
		if filter.WithPOI {
			object, err = GetObjectFromTanglePOI(blockIdStr, l.POIHandler)
			if err != nil {
				return err
			}
		} else {
			object.Block = block
		}
		err = l.Storage.UploadObject(blockIdStr, filter.BucketName, object, ctx)
		if err != nil {
			err = fmt.Errorf("can't upload the block '%s', error: %w", blockIdStr, err)
			return err
		}
	}
	return nil
}

func getSubscribedSignedPayload(taggedData iotago.TaggedData, expectedPublicKey []byte) (*datapayloads.SignedDataContainer, error) {
	// try to get signed data container from bytes
	signedPayload, err := datapayloads.NewSignedDataContainerFromBytes(taggedData.Data)
	if err != nil {
		return signedPayload, err
	}

	// get signed data public key
	publicKey, err := signedPayload.PublicKey()
	if err != nil {
		return signedPayload, err
	}

	// check if public keys are the same
	if !reflect.DeepEqual(publicKey, expectedPublicKey) {
		return signedPayload, err
	}

	return signedPayload, nil
}
