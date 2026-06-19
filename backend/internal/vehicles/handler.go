// Package vehicles
package vehicles

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/ize-302/beacon/backend/internal/common"
)

type VehicleHandler struct {
	APIGroup       *huma.Group
	VehicleService *VehicleService
}

func NewVehicleHandler(apiGroup *huma.Group, vehicleService *VehicleService) *VehicleHandler {
	return &VehicleHandler{APIGroup: apiGroup, VehicleService: vehicleService}
}

func (h *VehicleHandler) RegisterRoutes() {
	vehicleGroup := huma.NewGroup(h.APIGroup, "/vehicles")

	huma.Register(vehicleGroup, huma.Operation{
		OperationID:   "create-vehicle",
		Path:          "",
		Method:        http.MethodPost,
		Summary:       "Create new vehicles",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"Vehicles"},
	}, func(ctx context.Context, input *CreateVehicleRequest) (*common.BaseResponseBody[VehicleResponse], error) {
		return h.VehicleService.CreateVehicle(input)
	})

	huma.Register(vehicleGroup, huma.Operation{
		OperationID:   "get-vehicles",
		Path:          "",
		Method:        http.MethodGet,
		Summary:       "List out vehicles",
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
		DefaultStatus: http.StatusNoContent,
		Tags:          []string{"Vehicles"},
	}, func(ctx context.Context, input *DeleteVehicleParams) (*struct{}, error) {
		return h.VehicleService.DeleteVehicle(input)
	})
}
