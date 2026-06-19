// Package health
package health

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/ize-302/beacon/backend/internal/common"
)

type HealthHandler struct {
	APIGroup *huma.Group
}

func NewHealthHander(apiGroup *huma.Group) *HealthHandler {
	return &HealthHandler{APIGroup: apiGroup}
}

func (h *HealthHandler) RegisterRoutes() {
	healthGroup := huma.NewGroup(h.APIGroup, "/health")

	huma.Register(healthGroup, huma.Operation{
		OperationID:   "health-check",
		Path:          "",
		Method:        http.MethodGet,
		Summary:       "Health check",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"Health"},
	}, func(ctx context.Context, input *struct{}) (*common.BaseResponseBody[any], error) {
		resp := &common.BaseResponseBody[any]{}
		resp.Body.Data = nil
		resp.Body.Status = true
		resp.Body.Message = "All is well in Ba sing seh"
		return resp, nil
	})
}
