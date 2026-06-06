// Package models
package models

type Vehicle struct {
	ID          int    `json:"id"`
	PlateNumber string `json:"plate_number"`
}

type Rider struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type AssignmentRequest struct {
	ID        int `json:"id"`
	VehicleID int `json:"vehicle_id"`
	RiderID   int `json:"rider_id"`
}

type AssignmentResponse struct {
	ID      int      `json:"id"`
	Vehicle *Vehicle `json:"vehicle"`
	Rider   *Rider   `json:"rider"`
}

type Action string

const (
	CreateAction   Action = "create"
	DeleteAction   Action = "delete"
	AssignAction   Action = "assign"
	UnassignAction Action = "unassign"
)

type Entity string

const (
	RiderEntity      Entity = "rider"
	VehicleEntity    Entity = "vehicle"
	AssignmentEntity Entity = "assignment"
)

type AuditLogResponse struct {
	ID         int    `json:"id"`
	Action     Action `json:"action"`
	EntityType Entity `json:"entity_type"`
	EntityID   int    `json:"entity_id"`
	Timestamp  string `json:"timestamp"`
	Metadata   any    `json:"metadata"`
}
