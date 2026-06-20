// Package vehicles
package vehicles

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type VehicleType string

const (
	Car   VehicleType = "car"
	Bus   VehicleType = "bus"
	Truck VehicleType = "truck"
	Van   VehicleType = "van"
)

func (VehicleType) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: "string",
		Enum: []interface{}{string(Car), string(Bus), string(Truck), string(Van)},
	}
}

type Vehicle struct {
	ID          int
	PlateNumber string
	CreatedAt   time.Time
	VehicleType VehicleType
}
