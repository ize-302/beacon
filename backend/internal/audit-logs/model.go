// Package auditlogs
package auditlogs

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
