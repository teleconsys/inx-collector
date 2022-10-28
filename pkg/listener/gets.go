package listener

import (
	"collector/pkg/poi"
	"collector/pkg/storage"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
)

func GetTaggedDataFromId(blockId *inx.BlockId, client inx.INXClient, ctx context.Context) (string, *iotago.Block, error) {
	var block *iotago.Block
	rawBlock, err := client.ReadBlock(ctx, blockId)
	if err != nil {
		return "", block, err
	}
	block, err = rawBlock.UnwrapBlock(serializer.DeSeriModeNoValidation, &iotago.ProtocolParameters{})
	if err != nil {
		return "", block, err
	}
	blockPayload := block.Payload
	if blockPayload.PayloadType() != iotago.PayloadTaggedData {
		return "", block, nil
	}

	payloadBytes, _ := blockPayload.Serialize(serializer.DeSeriModeNoValidation, ctx)

	taggedData := iotago.TaggedData{}
	_, err = taggedData.Deserialize(payloadBytes, serializer.DeSeriModeNoValidation, ctx)
	if err != nil {
		return "", block, err
	}

	return string(taggedData.Tag), block, nil
}

func GetObjectFromTanglePOI(blockId string, poiHandler poi.POIHandler) (storage.Object, error) {
	var object storage.Object
	body, err := poiHandler.CreatePOI(blockId)
	if err != nil {
		return object, fmt.Errorf("can't get Proof of Inclusion for block '%s', error: %w", blockId, err)
	}
	object, err = storage.NewObject(body)
	if err != nil {
		return object, err
	}
	return object, nil
}

func GetObjectFromTangleBlock(blockId string, client inx.INXClient, ctx context.Context) (storage.Object, error) {
	var err error
	var object storage.Object
	var blockID inx.BlockId

	blockID.Id, err = hex.DecodeString(blockId)
	if err != nil {
		return object, err
	}

	block, err := client.ReadBlock(ctx, &blockID)
	if err != nil {
		return object, err
	}
	uBlock, err := block.UnwrapBlock(serializer.DeSeriModeNoValidation, &iotago.ProtocolParameters{})
	if err != nil {
		return object, err
	}

	object.Block = uBlock

	return object, nil
}
