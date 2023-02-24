package listener

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/datapayloads.go"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/builder"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/stretchr/testify/require"
)

// TODO: make nice tests

func TestSignedPayloadAnalysis(t *testing.T) {
	privateHexa := "5de46a7ba1bb8b58077e2201e555bd9827c21f949322bca26c5c44299fa835e87a882de7592ad1d6af7d19153b964f35891e2bdbc2e56beea659222b679781cc"
	publicHexa := "7a882de7592ad1d6af7d19153b964f35891e2bdbc2e56beea659222b679781cc"

	private, err := hex.DecodeString(privateHexa)
	if err != nil {
		panic(err)
	}

	public, err := hex.DecodeString(publicHexa)
	if err != nil {
		panic(err)
	}

	signer := datapayloads.NewInMemorySigner(ed25519.PrivateKey(private))

	signedDataContainer, err := datapayloads.NewSignedDataContainer(signer, []byte("test data"))
	if err != nil {
		panic(err)
	}

	dataPubKey, err := signedDataContainer.PublicKey()
	if err != nil {
		panic(err)
	}

	require.Equal(t, fmt.Sprintf("%v", dataPubKey), fmt.Sprintf("%v", public))
	require.NoError(t, signedDataContainer.VerifySignature())
}

func TestSendSignedPayloadKeyA(t *testing.T) {
	// PUBLIC KEY : 7a882de7592ad1d6af7d19153b964f35891e2bdbc2e56beea659222b679781cc
	privateHexa := "5de46a7ba1bb8b58077e2201e555bd9827c21f949322bca26c5c44299fa835e87a882de7592ad1d6af7d19153b964f35891e2bdbc2e56beea659222b679781cc"
	data := "test data"
	tag := "test tag"

	sendDataPayload(privateHexa, data, tag)
}

func TestSendSignedPayloadKeyB(t *testing.T) {
	// PUBLIC KEY : 6f1581709bb7b1ef030d210db18e3b0ba1c776fba65d8cdaad05415142d189f8
	privateHexa := "52fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c6496f1581709bb7b1ef030d210db18e3b0ba1c776fba65d8cdaad05415142d189f8"
	data := "test data"
	tag := "test tag"

	sendDataPayload(privateHexa, data, tag)
}

func TestSendWrongSignedPayloadKeyA(t *testing.T) {
	// PUBLIC KEY : 7a882de7592ad1d6af7d19153b964f35891e2bdbc2e56beea659222b679781cc
	privateHexa := "5de46a7ba1bb8b58077e2201e555bd9827c21f949322bca26c5c44299fa835e87a882de7592ad1d6af7d19153b964f35891e2bdbc2e56beea659222b679781cc"
	data := "test data"
	tag := "test tag"

	sendWrongDataPayload(privateHexa, data, tag)
}

func TestSendWrongSignedPayloadKeyB(t *testing.T) {
	// PUBLIC KEY : 6f1581709bb7b1ef030d210db18e3b0ba1c776fba65d8cdaad05415142d189f8
	privateHexa := "52fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c6496f1581709bb7b1ef030d210db18e3b0ba1c776fba65d8cdaad05415142d189f8"
	data := "test data"
	tag := "test tag"

	sendWrongDataPayload(privateHexa, data, tag)
}

func sendDataPayload(privateHexa, data, tag string) {
	private, err := hex.DecodeString(privateHexa)
	if err != nil {
		panic(err)
	}

	signer := datapayloads.NewInMemorySigner(ed25519.PrivateKey(private))

	signedDataContainer, err := datapayloads.NewSignedDataContainer(signer, []byte(data))
	if err != nil {
		panic(err)
	}

	signedData, err := signedDataContainer.Serialize(serializer.DeSeriModePerformValidation, nil)
	if err != nil {
		panic(err)
	}

	// create a new node API client
	nodeHTTPAPIClient := nodeclient.New("https://api.shimmer.network")

	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunc()

	// fetch the node's info to know the min. required PoW score
	info, err := nodeHTTPAPIClient.Info(ctx)
	if err != nil {
		panic(err)
	}

	taggedDataPayload := &iotago.TaggedData{
		Tag:  []byte(tag),
		Data: signedData,
	}

	// get some tips from the node
	tipsResponse, err := nodeHTTPAPIClient.Tips(ctx)
	if err != nil {
		panic(err)
	}

	tips, err := tipsResponse.Tips()
	if err != nil {
		panic(err)
	}

	// get the current protocol parameters
	protoParas := info.Protocol

	// build a block by adding the paylod and the tips and then do local Proof-of-Work
	block, err := builder.NewBlockBuilder().
		Payload(taggedDataPayload).
		Parents(tips).
		ProofOfWork(ctx, nil, float64(protoParas.MinPoWScore)).
		Build()
	if err != nil {
		panic(err)
	}

	// submit the block to the node
	b, err := nodeHTTPAPIClient.SubmitBlock(ctx, block, nil)
	if err != nil {
		panic(err)
	}

	id, err := b.ID()
	if err != nil {
		panic(err)
	}
	fmt.Println(id.ToHex())
}

func sendWrongDataPayload(privateHexa, data, tag string) {
	private, err := hex.DecodeString(privateHexa)
	if err != nil {
		panic(err)
	}

	signer := datapayloads.NewInMemorySigner(ed25519.PrivateKey(private))

	signedDataContainer, err := datapayloads.NewSignedDataContainer(signer, []byte(data))
	if err != nil {
		panic(err)
	}

	_, err = signedDataContainer.Signature.Deserialize([]byte("52fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c6496f1581709bb7b1ef030d210db18e3b0ba1c776fba65d8cdaad05415142d189f8"), serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic(err)
	}

	signedData, err := signedDataContainer.Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		panic(err)
	}

	// create a new node API client
	nodeHTTPAPIClient := nodeclient.New("https://api.shimmer.network")

	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunc()

	// fetch the node's info to know the min. required PoW score
	info, err := nodeHTTPAPIClient.Info(ctx)
	if err != nil {
		panic(err)
	}

	taggedDataPayload := &iotago.TaggedData{
		Tag:  []byte(tag),
		Data: signedData,
	}

	// get some tips from the node
	tipsResponse, err := nodeHTTPAPIClient.Tips(ctx)
	if err != nil {
		panic(err)
	}

	tips, err := tipsResponse.Tips()
	if err != nil {
		panic(err)
	}

	// get the current protocol parameters
	protoParas := info.Protocol

	// build a block by adding the paylod and the tips and then do local Proof-of-Work
	block, err := builder.NewBlockBuilder().
		Payload(taggedDataPayload).
		Parents(tips).
		ProofOfWork(ctx, nil, float64(protoParas.MinPoWScore)).
		Build()
	if err != nil {
		panic(err)
	}

	// submit the block to the node
	b, err := nodeHTTPAPIClient.SubmitBlock(ctx, block, nil)
	if err != nil {
		panic(err)
	}

	id, err := b.ID()
	if err != nil {
		panic(err)
	}
	fmt.Println(id.ToHex())
}

func TestGetVerifyDataPayload(t *testing.T) {
	nodeHTTPAPIClient := nodeclient.New("https://api.shimmer.network")

	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunc()

	// fetch the node's info to know the min. required PoW score
	_, err := nodeHTTPAPIClient.Info(ctx)
	if err != nil {
		panic(err)
	}

	id, err := iotago.BlockIDFromHexString("0x7afd1bcffe23f3b6e7e7119436de47d2a95042d5aabd84049df7adf8aa968c59")
	if err != nil {
		panic(err)
	}

	block, err := nodeHTTPAPIClient.BlockByBlockID(ctx, id, nil)
	if err != nil {
		panic(err)
	}

	payloadBytes, _ := block.Payload.Serialize(serializer.DeSeriModePerformValidation, ctx)

	taggedData := iotago.TaggedData{}
	_, err = taggedData.Deserialize(payloadBytes, serializer.DeSeriModePerformValidation, ctx)
	if err != nil {
		panic(err)
	}

	sdc, err := datapayloads.NewSignedDataContainerFromBytes(taggedData.Data)
	if err != nil {
		panic(err)
	}

	require.NoError(t, sdc.VerifySignature())
}

func TestGetVerifyWrongDataPayload(t *testing.T) {
	nodeHTTPAPIClient := nodeclient.New("https://api.shimmer.network")

	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunc()

	// fetch the node's info to know the min. required PoW score
	_, err := nodeHTTPAPIClient.Info(ctx)
	if err != nil {
		panic(err)
	}

	id, err := iotago.BlockIDFromHexString("0x2b5520706f268f680fd15789aab222cd89c5496d54f5433f00e7980d0b78705f")
	if err != nil {
		panic(err)
	}

	block, err := nodeHTTPAPIClient.BlockByBlockID(ctx, id, nil)
	if err != nil {
		panic(err)
	}

	payloadBytes, _ := block.Payload.Serialize(serializer.DeSeriModePerformValidation, ctx)

	taggedData := iotago.TaggedData{}
	_, err = taggedData.Deserialize(payloadBytes, serializer.DeSeriModePerformValidation, ctx)
	if err != nil {
		panic(err)
	}

	sdc, err := datapayloads.NewSignedDataContainerFromBytes(taggedData.Data)
	if err != nil {
		panic(err)
	}

	require.Error(t, sdc.VerifySignature())
}
