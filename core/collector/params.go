package collector

import (
	"collector/pkg/api"
	"collector/pkg/listener"
	"collector/pkg/poi"
	"collector/pkg/storage"

	"github.com/iotaledger/hive.go/core/app"
)

var ParamsListener = &listener.Parameters{}
var ParamsStorage = &storage.Parameters{}
var ParamsRestAPI = &api.Parameters{}
var ParamsPOI = &poi.Parameters{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"listener": ParamsListener,
		"POI":      ParamsPOI,
		"restAPI":  ParamsRestAPI,
		"storage":  ParamsStorage,
	},
	Masked: nil,
}
