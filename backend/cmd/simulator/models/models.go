// Package models
package models

type Route struct {
	Name        string      `json:"name"`
	Coordinates [][]float64 `json:"coordinates"`
}

type Gps struct {
	ID           int         `json:"id"`
	GpsID        int         `json:"gps_id"`
	CurrentIndex int         `json:"current_index"`
	Routes       [][]float64 `json:"routes"`
}

type GpsPayload struct {
	GpsID     int     `json:"gps_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
