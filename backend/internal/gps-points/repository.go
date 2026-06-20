package gpspoints

import (
	"database/sql"

	_ "embed"
)

//go:embed queries/insert_gpspoint.sql
var insertGpsPoint string

//go:embed queries/select_gpspoints.sql
var selectGpsPoints string

type GpsPointRepository struct {
	db *sql.DB
}

func NewGpsPointRepository(db *sql.DB) *GpsPointRepository {
	return &GpsPointRepository{db: db}
}

func (r *GpsPointRepository) SaveGpsPointRepo(input *CreateGpsPointRequest) (*GpsPointResponse, error) {
	var gpspoint GpsPointResponse
	err := r.db.QueryRow(insertGpsPoint, input.Body.GpsID, input.Body.Bearing, input.Body.Latitude, input.Body.Longitude, input.Body.Timestamp).Scan(
		&gpspoint.ID, &gpspoint.GpsID, &gpspoint.Bearing, &gpspoint.Latitude, &gpspoint.Longitude, &gpspoint.Timestamp, &gpspoint.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &gpspoint, nil
}

func (r *GpsPointRepository) FetchGpsPointsRepo() (*[]GpsPointResponse, error) {
	rows, err := r.db.Query(selectGpsPoints)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gpspoints []GpsPointResponse
	for rows.Next() {
		var gpspoint GpsPointResponse
		// select_gpspoints.sql does not include bearing
		if err = rows.Scan(&gpspoint.ID, &gpspoint.GpsID, &gpspoint.Latitude, &gpspoint.Longitude, &gpspoint.Timestamp, &gpspoint.CreatedAt); err != nil {
			return nil, err
		}
		gpspoints = append(gpspoints, gpspoint)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &gpspoints, nil
}
