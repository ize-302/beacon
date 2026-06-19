package vehicles

import (
	"database/sql"

	_ "embed"

	"github.com/danielgtaylor/huma/v2"
)

//go:embed queries/insert_vehicle.sql
var insertVehicle string

//go:embed queries/select_vehicles.sql
var selectVehicles string

//go:embed queries/select_vehicle.sql
var selectVehicle string

//go:embed queries/delete_vehicle.sql
var deleteVehicle string

type VehicleRepository struct {
	db *sql.DB
}

func NewVehicleRepository(db *sql.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

func (r *VehicleRepository) CreateVehicleRepo(input *CreateVehicleRequest) (*Vehicle, error) {
	var vehicle Vehicle
	err := r.db.QueryRow(insertVehicle, input.Body.PlateNumber, input.Body.VehicleType).Scan(&vehicle.ID, &vehicle.PlateNumber, &vehicle.VehicleType, &vehicle.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *VehicleRepository) FetchVehiclesRepo() (*[]Vehicle, error) {
	var vehicles []Vehicle
	rows, err := r.db.Query(selectVehicles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var vehicle Vehicle
		err = rows.Scan(&vehicle.ID, &vehicle.PlateNumber, &vehicle.VehicleType, &vehicle.CreatedAt)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, vehicle)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &vehicles, nil
}

func (r *VehicleRepository) getVehicleByIDRepo(id int) *sql.Row {
	row := r.db.QueryRow(selectVehicle, id)
	return row
}

func (r *VehicleRepository) FetchVehicleRepo(input *GetVehicleParams) (*Vehicle, error) {
	var vehicle Vehicle
	row := r.getVehicleByIDRepo(input.ID)
	switch err := row.Scan(&vehicle.ID, &vehicle.PlateNumber, &vehicle.VehicleType, &vehicle.CreatedAt); err {
	case sql.ErrNoRows:
		return nil, huma.Error404NotFound("vehicle not found", err)
	case nil:
		return &vehicle, nil
	default:
		panic(err)
	}
}

func (r *VehicleRepository) DeleteVehicleRepo(input *DeleteVehicleParams) error {
	row := r.db.QueryRow(`SELECT id FROM vehicles WHERE id = $1`, input.ID)
	switch err := row.Scan(&input.ID); err {
	case sql.ErrNoRows:
		return huma.Error404NotFound("vehicle not found", err)
	case nil:
		_ = r.db.QueryRow(deleteVehicle, input.ID)
		return nil
	default:
		panic(err)
	}
}
