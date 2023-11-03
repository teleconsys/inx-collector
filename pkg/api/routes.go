package api

import (
	"collector/pkg/listener"
	"collector/pkg/storage"
	"encoding/json"
	"fmt"
	"net/http"

	"strings"

	"github.com/iotaledger/inx-app/pkg/httpserver"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/labstack/echo/v4"
)

const (
	// ParameterBlockID is used to identify a block by its ID.
	ParameterBlockID = "blockId"
	// ParameterWithPOI is used to identify wether a get request should recover also the POI.
	ParameterWithPOI = "withPOI"
	// ParameterPrivate is used to identify wether a get request should use a private bucket.
	ParameterPrivate = "private"
	// ParameterUserId is used to identify the user id for a private request.
	ParameterUserId = "userId"
	// ParameterBucketName is used to identify the user's bucket name fora private request.
	ParameterBucketName = "bucketName"
	// ParameterFilterId is used to identify the filter id
	ParameterFilterId = "filterId"
	// ParameterLifecycleDays is used to express the number of days before data expiration in the bucket.
	ParameterLifecycleDays = "days"

	RouteGetBlock     = "/block/:" + ParameterBlockID
	RouteDeleteBlock  = "/block/:" + ParameterBlockID
	RouteStore        = "/block"
	RouteSubscribe    = "/filter"
	RouteUnsubscribe  = "/filter/:" + ParameterFilterId
	RouteCreateBucket = "/bucket"
)

func (s *Server) setupRoutes(e *echo.Echo) {
	e.GET(RouteGetBlock, func(c echo.Context) error {
		var err error
		s.apiLogStart(RouteGetBlock)
		defer s.apiLogEnd(RouteGetBlock, err)

		params, err := s.parseObjectInput(c)
		if err != nil {
			return httpserver.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("%v", err))
		}
		if params.WithPOI {
			resp, err := s.getBlockWithPOI(params.BlockId, params.BucketName, c)
			if err != nil {
				return httpserver.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("%v", err))
			}
			return httpserver.JSONResponse(c, http.StatusOK, &resp)
		}

		resp, err := s.getBlock(params.BlockId, params.BucketName, c)
		if err != nil {
			return httpserver.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("%v", err))
		}
		return httpserver.JSONResponse(c, http.StatusOK, &resp)
	})
	e.POST(RouteStore, func(c echo.Context) error {
		var err error
		s.apiLogStart(RouteStore)
		defer s.apiLogEnd(RouteStore, err)

		blockId, bucketName, err := s.storeBlockFromTangle(c)
		if err != nil {
			return httpserver.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("%v", err))
		}
		return httpserver.JSONResponse(c, http.StatusOK, fmt.Sprintf("Block '%s' uploaded to bucket '%s'", blockId, bucketName))
	})
	e.POST(RouteSubscribe, func(c echo.Context) error {
		var err error
		s.apiLogStart(RouteSubscribe)
		defer s.apiLogEnd(RouteSubscribe, err)

		filterId, tag, err := s.subscribeToTag(c)
		if err != nil {
			return httpserver.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("%v", err))
		}
		return httpserver.JSONResponse(c, http.StatusOK, fmt.Sprintf("Subscription to '%s' started, id is: '%s'", tag, filterId))
	})
	e.POST(RouteCreateBucket, func(c echo.Context) error {
		var err error
		s.apiLogStart(RouteCreateBucket)
		defer s.apiLogEnd(RouteCreateBucket, err)

		bucketName, err := s.createBucketFromRequest(c)
		if err != nil {
			return httpserver.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("could not create bucket, error: %v", err))
		}
		return httpserver.JSONResponse(c, http.StatusOK, fmt.Sprintf("Bucket '%s' created", bucketName))
	})
	e.DELETE(RouteDeleteBlock, func(c echo.Context) error {
		var err error
		s.apiLogStart(RouteDeleteBlock)
		defer s.apiLogEnd(RouteDeleteBlock, err)

		params, err := s.parseObjectInput(c)
		if err != nil {
			return httpserver.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("%v", err))
		}

		err = s.Collector.Storage.DeleteObject(params.BucketName, params.BlockId, s.Context)
		if err != nil {
			return httpserver.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("%v", err))
		}

		return httpserver.JSONResponse(c, http.StatusOK, fmt.Sprintf("Object '%s' removed from bucket '%s'", params.BlockId, params.BucketName))
	})
	e.DELETE(RouteUnsubscribe, func(c echo.Context) error {
		var err error
		s.apiLogStart(RouteUnsubscribe)
		defer s.apiLogEnd(RouteUnsubscribe, err)

		filterId := strings.ToLower(c.Param(ParameterFilterId))
		s.Collector.Listener.RemoveFilter(filterId)

		return httpserver.JSONResponse(c, http.StatusOK, fmt.Sprintf("Subscription with id '%s' has stopped", filterId))
	})
}

func (s *Server) getBlock(blockId string, bucketName string, c echo.Context) (*iotago.Block, error) {
	object, err := s.getObjectFromStorage(blockId, bucketName)
	if err != nil {
		return nil, err
	}
	return object.Block, nil
}

func (s *Server) getBlockWithPOI(blockId string, bucketName string, c echo.Context) (storage.Object, error) {
	object, err := s.getObjectFromStorage(blockId, bucketName)
	if err != nil {
		return storage.Object{}, err
	}

	if object.Milestone == nil || object.Proof == nil {
		return storage.Object{}, fmt.Errorf("error: malformed or missing proof of inclusion")
	}
	return storage.Object{Milestone: object.Milestone, Block: object.Block, Proof: object.Proof}, nil
}

func (s *Server) getObjectFromStorage(blockId string, bucketName string) (storage.Object, error) {
	var object storage.Object
	resp, err := s.Collector.Storage.GetObject(bucketName, blockId, s.Context)
	if err != nil {
		return object, err
	}

	err = json.NewDecoder(resp).Decode(&object)
	if err != nil {
		return object, err
	}
	return object, nil
}

func (s *Server) storeBlockFromTangle(c echo.Context) (string, string, error) {
	var request RequestStoreBody
	err := extractRequestBody(&request, c)
	if err != nil {
		return "", "", err
	}

	bucketName := s.Collector.Storage.DefaultBucketName
	if request.BucketName != "" {
		bucketName = request.BucketName
	}

	var object storage.Object
	if request.WithPOI {
		object, err = listener.GetObjectFromTanglePOI(request.BlockId, s.Collector.POIHandler)
	} else {
		object, err = listener.GetObjectFromTangleBlock(request.BlockId, s.Collector.NodeBridge.Client(), s.Context)
	}
	if err != nil {
		return "", "", err
	}

	err = s.Collector.Storage.UploadObject(request.BlockId, bucketName, object, s.Context)
	if err != nil {
		return "", "", err
	}

	return request.BlockId, bucketName, nil
}

func (s *Server) subscribeToTag(c echo.Context) (string, string, error) {
	var request RequestSubscribeBody
	err := extractRequestBody(&request, c)
	if err != nil {
		return "", "", err
	}

	bucketName := s.Collector.Storage.DefaultBucketName
	if request.BucketName != "" {
		bucketName = request.BucketName
	}

	filter, err := listener.NewFilter(request.Tag, request.PublicKey, bucketName, request.Duration, request.WithPOI)
	if err != nil {
		return "", "", err
	}

	filterId, err := s.Collector.Listener.AddFilter(filter)
	if err != nil {
		return "", "", err
	}

	return filterId, request.Tag, nil
}

func (s *Server) createBucketFromRequest(c echo.Context) (string, error) {
	var request RequestCreateBucket
	err := extractRequestBody(&request, c)
	if err != nil {
		return "", err
	}

	err = s.Collector.Storage.CreateBucket(request.BucketName, s.Context)
	if err != nil {
		return "", err
	}

	if request.LifecycleDays != 0 {
		err = s.Collector.Storage.SetBucketExpirationDays(request.BucketName, request.LifecycleDays, s.Context)
		if err != nil {
			return "", err
		}
	}

	return request.BucketName, nil
}
