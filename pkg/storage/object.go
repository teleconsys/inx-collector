package storage

import (
	"bytes"
	"encoding/json"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/merklehasher"
)

type Object struct {
	Milestone *iotago.Milestone   `json:"milestone,omitempty"`
	Block     *iotago.Block       `json:"block"`
	Proof     *merklehasher.Proof `json:"proof,omitempty"`
}

func NewObject(reader io.Reader) (Object, error) {
	var object Object

	b, err := io.ReadAll(reader)
	if err != nil {
		return object, err
	}

	err = json.Unmarshal(b, &object)
	if err != nil {
		return object, err
	}
	return object, nil
}

func (o *Object) GetByteReader() (*bytes.Reader, error) {
	var blockReader *bytes.Reader

	objectJson, err := json.Marshal(o)
	if err != nil {
		return blockReader, err
	}

	blockReader = bytes.NewReader(objectJson)
	return blockReader, nil
}
