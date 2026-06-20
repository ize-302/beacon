package gps

import (
	"database/sql"
	"time"

	_ "embed"

	"github.com/danielgtaylor/huma/v2"
	"github.com/ize-302/beacon/backend/internal/vehicles"
)

//go:embed queries/insert_gps.sql
var insertGps string

//go:embed queries/select_gps_devices.sql
var selectGpsDevices string

//go:embed queries/select_gps.sql
var selectGps string

//go:embed queries/delete_gps.sql
var deleteGps string

//go:embed queries/select_gps_history.sql
var getGpsHistory string

type GpsRepository struct {
	db *sql.DB
}

func NewGpsRepository(db *sql.DB) *GpsRepository {
	return &GpsRepository{db: db}
}

func scanGpsRow(row *sql.Row) (*GpsResponse, error) {
	var gps GpsResponse
	gps.Vehicle = &vehicles.VehicleResponse{}
	var lat, lng sql.NullFloat64
	var lastAt sql.NullTime
	switch err := row.Scan(&gps.ID, &gps.SN, &gps.CreatedAt, &gps.Vehicle.ID, &gps.Vehicle.PlateNumber, &gps.Vehicle.CreatedAt, &lat, &lng, &lastAt); err {
	case sql.ErrNoRows:
		return nil, huma.Error404NotFound("gps device not found", err)
	case nil:
		if lat.Valid && lng.Valid {
			updatedAt := time.Time{}
			if lastAt.Valid {
				updatedAt = lastAt.Time
			}
			gps.LastCoordinate = &Coordinate{
				Latitude:  lat.Float64,
				Longitude: lng.Float64,
				UpdatedAt: updatedAt,
			}
		}
		return &gps, nil
	default:
		panic(err)
	}
}

func (r *GpsRepository) CreateGpsRepo(input *CreateGpsRequest) (*GpsResponse, error) {
	var gps GpsResponse
	gps.Vehicle = &vehicles.VehicleResponse{}
	err := r.db.QueryRow(insertGps, input.Body.SN, input.Body.VehicleID).Scan(
		&gps.ID, &gps.SN, &gps.CreatedAt,
		&gps.Vehicle.ID, &gps.Vehicle.PlateNumber, &gps.Vehicle.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &gps, nil
}

func (r *GpsRepository) FetchGpsDevicesRepo() (*[]GpsResponse, error) {
	rows, err := r.db.Query(selectGpsDevices)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gpsDevices []GpsResponse
	for rows.Next() {
		var gps GpsResponse
		gps.Vehicle = &vehicles.VehicleResponse{}
		var lat, lng sql.NullFloat64
		var lastAt sql.NullTime
		if err = rows.Scan(&gps.ID, &gps.SN, &gps.CreatedAt, &gps.Vehicle.ID, &gps.Vehicle.PlateNumber, &gps.Vehicle.CreatedAt, &lat, &lng, &lastAt); err != nil {
			return nil, err
		}
		if lat.Valid && lng.Valid {
			updatedAt := time.Time{}
			if lastAt.Valid {
				updatedAt = lastAt.Time
			}
			gps.LastCoordinate = &Coordinate{
				Latitude:  lat.Float64,
				Longitude: lng.Float64,
				UpdatedAt: updatedAt,
			}
		}
		gpsDevices = append(gpsDevices, gps)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &gpsDevices, nil
}

func (r *GpsRepository) FetchGpsRepo(input *GetGpsParams) (*GpsResponse, error) {
	row := r.db.QueryRow(selectGps, input.ID)
	return scanGpsRow(row)
}

func (r *GpsRepository) DeleteGpsRepo(input *DeleteGpsParams) error {
	row := r.db.QueryRow(`SELECT id FROM gps_devices WHERE id = $1`, input.ID)
	switch err := row.Scan(&input.ID); err {
	case sql.ErrNoRows:
		return huma.Error404NotFound("gps device not found", err)
	case nil:
		_ = r.db.QueryRow(deleteGps, input.ID)
		return nil
	default:
		panic(err)
	}
}

func (r *GpsRepository) FetchGpsHistoryRepo(input *GetGpsHistoryParams) (*GpsHistoryResponse, error) {
	rows, err := r.db.Query(getGpsHistory, input.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gpsHistory GpsHistoryResponse
	coordinates := []Coordinate{}
	found := false

	for rows.Next() {
		var lat, lng sql.NullFloat64
		if err := rows.Scan(&gpsHistory.GpsID, &gpsHistory.GpsSN, &lat, &lng); err != nil {
			panic(err)
		}
		found = true
		if lat.Valid && lng.Valid {
			coordinates = append(coordinates, Coordinate{
				Latitude:  lat.Float64,
				Longitude: lng.Float64,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if !found {
		return nil, huma.Error404NotFound("gps device not found", nil)
	}
	gpsHistory.Coordinates = &coordinates
	return &gpsHistory, nil
}
