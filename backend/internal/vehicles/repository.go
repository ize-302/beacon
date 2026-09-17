package vehicles

import (
	"database/sql"
	"time"

	_ "embed"

	"github.com/danielgtaylor/huma/v2"
)

//go:embed queries/insert_vehicle.sql
var insertVehicle string

//go:embed queries/select_vehicles.sql
var selectVehicles string

//go:embed queries/select_vehicle.sql
var selectVehicle string

//go:embed queries/select_vehicle_history.sql
var selectVehicleHistory string

//go:embed queries/delete_vehicle.sql
var deleteVehicle string

type VehicleRepository struct {
	db *sql.DB
}

func NewVehicleRepository(db *sql.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

// scanVehicle reads the shared column list used by select_vehicle(s).sql.
func scanVehicle(s scanner) (*VehicleResponse, error) {
	var v VehicleResponse
	var lat, lng sql.NullFloat64
	var lastAt sql.NullTime

	if err := s.Scan(
		&v.ID, &v.PlateNumber, &v.VehicleType, &v.DeviceSN, &v.CreatedAt,
		&lat, &lng, &lastAt,
	); err != nil {
		return nil, err
	}

	if lat.Valid && lng.Valid {
		updatedAt := time.Time{}
		if lastAt.Valid {
			updatedAt = lastAt.Time
		}
		v.LastCoordinate = &Coordinate{
			Latitude:  lat.Float64,
			Longitude: lng.Float64,
			UpdatedAt: updatedAt,
		}
	}
	return &v, nil
}

func (r *VehicleRepository) CreateVehicleRepo(input *CreateVehicleRequest) (*VehicleResponse, error) {
	var v VehicleResponse
	err := r.db.QueryRow(insertVehicle, input.Body.PlateNumber, input.Body.VehicleType, input.Body.DeviceSN).
		Scan(&v.ID, &v.PlateNumber, &v.VehicleType, &v.DeviceSN, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VehicleRepository) FetchVehiclesRepo() ([]VehicleResponse, error) {
	rows, err := r.db.Query(selectVehicles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vehicles := []VehicleResponse{}
	for rows.Next() {
		v, err := scanVehicle(rows)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, *v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return vehicles, nil
}

func (r *VehicleRepository) FetchVehicleRepo(input *GetVehicleParams) (*VehicleResponse, error) {
	v, err := scanVehicle(r.db.QueryRow(selectVehicle, input.ID))
	if err == sql.ErrNoRows {
		return nil, huma.Error404NotFound("vehicle not found", err)
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *VehicleRepository) DeleteVehicleRepo(input *DeleteVehicleParams) error {
	res, err := r.db.Exec(deleteVehicle, input.ID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return huma.Error404NotFound("vehicle not found", nil)
	}
	return nil
}

func (r *VehicleRepository) FetchVehicleHistoryRepo(input *GetVehicleHistoryParams) (*VehicleHistoryResponse, error) {
	rows, err := r.db.Query(selectVehicleHistory, input.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history VehicleHistoryResponse
	coordinates := []Coordinate{}
	found := false

	for rows.Next() {
		var lat, lng sql.NullFloat64
		if err := rows.Scan(&history.VehicleID, &history.PlateNumber, &lat, &lng); err != nil {
			return nil, err
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
		return nil, huma.Error404NotFound("vehicle not found", nil)
	}

	history.Coordinates = &coordinates
	return &history, nil
}
