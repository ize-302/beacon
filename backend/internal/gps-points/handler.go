// Package gpspoints
package gpspoints

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/ize-302/beacon/backend/internal/common"
)

type GpsPointHandler struct {
	APIGroup       *huma.Group
	GpsPointService *GpsPointService
}

func NewGpsPointHandler(apiGroup *huma.Group, gpsPointService *GpsPointService) *GpsPointHandler {
	return &GpsPointHandler{APIGroup: apiGroup, GpsPointService: gpsPointService}
}

func (h *GpsPointHandler) RegisterRoutes() {
	gpsPointGroup := huma.NewGroup(h.APIGroup, "/gps-points")

	huma.Register(gpsPointGroup, huma.Operation{
		OperationID:   "create-gps-point",
		Path:          "",
		Method:        http.MethodPost,
		Summary:       "Record a GPS point",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"GPS Points"},
	}, func(ctx context.Context, input *CreateGpsPointRequest) (*common.BaseResponseBody[GpsPointResponse], error) {
		return h.GpsPointService.SaveGpsPoint(input)
	})

	huma.Register(gpsPointGroup, huma.Operation{
		OperationID:   "get-gps-points",
		Path:          "",
		Method:        http.MethodGet,
		Summary:       "List GPS points",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"GPS Points"},
	}, func(ctx context.Context, input *struct{}) (*common.BaseResponseBody[[]GpsPointResponse], error) {
		return h.GpsPointService.FetchGpsPoints()
	})
}
