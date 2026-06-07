// Package models
package models

type Route struct {
	Name        string      `json:"name"`
	Coordinates [][]float64 `json:"coordinates"`
}

type Gps struct {
	ID           int         `json:"id"`
	AssignmentID int         `json:"assignment_id"`
	CurrentIndex int         `json:"current_index"`
	Routes       [][]float64 `json:"routes"`
}

type GpsPayload struct {
	AssignmentID int     `json:"assignment_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}
