// Package vehicles
package vehicles

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/ize-302/beacon/backend/internal/common"
)

type VehicleHandler struct {
	APIGroup       *huma.Group
	VehicleService *VehicleService
	Router         chi.Router
}

func NewVehicleHandler(apiGroup *huma.Group, vehicleService *VehicleService, router chi.Router) *VehicleHandler {
	return &VehicleHandler{APIGroup: apiGroup, VehicleService: vehicleService, Router: router}
}

func (h *VehicleHandler) RegisterRoutes() {
	vehicleGroup := huma.NewGroup(h.APIGroup, "/vehicles")

	// Registered on the router directly: SSE is a streaming response, which does
	// not fit huma's request/response model.
	h.Router.Get("/api/v1/vehicles/events", h.streamNewVehicles)

	huma.Register(vehicleGroup, huma.Operation{
		OperationID:   "create-vehicle",
		Path:          "",
		Method:        http.MethodPost,
		Summary:       "Create a vehicle",
		Description:   "The vehicle is tracked from the moment it exists; there is no separate device to register.",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"Vehicles"},
	}, func(ctx context.Context, input *CreateVehicleRequest) (*common.BaseResponseBody[VehicleResponse], error) {
		return h.VehicleService.CreateVehicle(input)
	})

	huma.Register(vehicleGroup, huma.Operation{
		OperationID:   "get-vehicles",
		Path:          "",
		Method:        http.MethodGet,
		Summary:       "List vehicles",
		Description:   "Includes each vehicle's last known coordinate, if it has one.",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"Vehicles"},
	}, func(ctx context.Context, input *struct{}) (*common.BaseResponseBody[[]VehicleResponse], error) {
		return h.VehicleService.FetchVehicles()
	})

	huma.Register(vehicleGroup, huma.Operation{
		OperationID:   "get-single-vehicle",
		Path:          "/{id}",
		Method:        http.MethodGet,
		Summary:       "Get vehicle",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"Vehicles"},
	}, func(ctx context.Context, input *GetVehicleParams) (*common.BaseResponseBody[VehicleResponse], error) {
		return h.VehicleService.FetchVehicle(input)
	})

	huma.Register(vehicleGroup, huma.Operation{
		OperationID:   "delete-vehicle",
		Path:          "/{id}",
		Method:        http.MethodDelete,
		Summary:       "Delete vehicle",
		Description:   "Removes the vehicle and its recorded history.",
		DefaultStatus: http.StatusNoContent,
		Tags:          []string{"Vehicles"},
	}, func(ctx context.Context, input *DeleteVehicleParams) (*struct{}, error) {
		return h.VehicleService.DeleteVehicle(input)
	})

	huma.Register(vehicleGroup, huma.Operation{
		OperationID:   "get-vehicle-history",
		Path:          "/{id}/history",
		Method:        http.MethodGet,
		Summary:       "Get vehicle location history",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"Vehicles"},
	}, func(ctx context.Context, input *GetVehicleHistoryParams) (*common.BaseResponseBody[VehicleHistoryResponse], error) {
		return h.VehicleService.FetchVehicleHistory(input)
	})
}

func (h *VehicleHandler) streamNewVehicles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := h.VehicleService.EventHub.Subscribe()
	defer h.VehicleService.EventHub.Unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case vehicle := <-ch:
			data, err := json.Marshal(vehicle)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
