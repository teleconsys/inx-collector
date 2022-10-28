package poi

import (
	"fmt"
	"io"
	"net/http"
)

type POIHandler struct {
	APIUrl string
}

func NewPOIHandler(params Parameters) POIHandler {
	var apiUrl string
	if params.IsPlugin {
		apiUrl = params.HostUrl + "/create/"
	} else {
		apiUrl = params.HostUrl + "/api/poi/v1/create/"
	}

	return POIHandler{APIUrl: apiUrl}
}

func (poi *POIHandler) CreatePOI(id string) (io.ReadCloser, error) {
	resp, err := http.Get(poi.APIUrl + "0x" + id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("POI request failed, status: %s", resp.Status)
		return nil, err
	}
	return resp.Body, nil
}
