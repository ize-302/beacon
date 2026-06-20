// Package gps
package gps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/ize-302/beacon/backend/internal/common"
)

type GpsHandler struct {
	APIGroup   *huma.Group
	GpsService *GpsService
	Router     chi.Router
}

func NewGpsHandler(apiGroup *huma.Group, gpsService *GpsService, router chi.Router) *GpsHandler {
	return &GpsHandler{APIGroup: apiGroup, GpsService: gpsService, Router: router}
}

func (h *GpsHandler) RegisterRoutes() {
	gpsGroup := huma.NewGroup(h.APIGroup, "/gps-devices")

	h.Router.Get("/api/v1/gps-devices/events", h.streamNewDevices)

	huma.Register(gpsGroup, huma.Operation{
		OperationID:   "create-gps-device",
		Path:          "",
		Method:        http.MethodPost,
		Summary:       "Register a GPS device",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"GPS Devices"},
	}, func(ctx context.Context, input *CreateGpsRequest) (*common.BaseResponseBody[GpsResponse], error) {
		return h.GpsService.CreateGps(input)
	})

	huma.Register(gpsGroup, huma.Operation{
		OperationID:   "get-gps-devices",
		Path:          "",
		Method:        http.MethodGet,
		Summary:       "List GPS devices",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"GPS Devices"},
	}, func(ctx context.Context, input *struct{}) (*common.BaseResponseBody[[]GpsResponse], error) {
		return h.GpsService.FetchGpsDevices()
	})

	huma.Register(gpsGroup, huma.Operation{
		OperationID:   "get-gps-device",
		Path:          "/{id}",
		Method:        http.MethodGet,
		Summary:       "Get a GPS device",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"GPS Devices"},
	}, func(ctx context.Context, input *GetGpsParams) (*common.BaseResponseBody[GpsResponse], error) {
		return h.GpsService.FetchGps(input)
	})

	huma.Register(gpsGroup, huma.Operation{
		OperationID:   "delete-gps-device",
		Path:          "/{id}",
		Method:        http.MethodDelete,
		Summary:       "Delete a GPS device",
		DefaultStatus: http.StatusNoContent,
		Tags:          []string{"GPS Devices"},
	}, func(ctx context.Context, input *DeleteGpsParams) (*struct{}, error) {
		return h.GpsService.DeleteGps(input)
	})

	huma.Register(gpsGroup, huma.Operation{
		OperationID:   "get-gps-history",
		Path:          "/{id}/history",
		Method:        http.MethodGet,
		Summary:       "Get GPS device location history",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"GPS Devices"},
	}, func(ctx context.Context, input *GetGpsHistoryParams) (*common.BaseResponseBody[GpsHistoryResponse], error) {
		return h.GpsService.FetchGpsHistory(input)
	})
}

func (h *GpsHandler) streamNewDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := h.GpsService.EventHub.Subscribe()
	defer h.GpsService.EventHub.Unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case gps := <-ch:
			data, err := json.Marshal(gps)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
