package gpspoints

import (
	"database/sql"

	_ "embed"

	"github.com/lib/pq"
)

//go:embed queries/insert_gpspoint.sql
var insertGpsPoint string

//go:embed queries/insert_gpspoints_batch.sql
var insertGpsPointsBatch string

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

// SaveGpsPointsRepo writes a whole batch in one round trip. The simulator used to
// send one point per request; at 500 vehicles that was 500 inserts a second, each
// with its own round trip and implicit transaction.
func (r *GpsPointRepository) SaveGpsPointsRepo(points []CreateGpsPoint) (int, error) {
	if len(points) == 0 {
		return 0, nil
	}

	gpsIDs := make(pq.Int64Array, len(points))
	bearings := make(pq.Float64Array, len(points))
	latitudes := make(pq.Float64Array, len(points))
	longitudes := make(pq.Float64Array, len(points))
	timestamps := make(pq.Int64Array, len(points))

	for i, p := range points {
		gpsIDs[i] = int64(p.GpsID)
		bearings[i] = p.Bearing
		latitudes[i] = p.Latitude
		longitudes[i] = p.Longitude
		timestamps[i] = p.Timestamp
	}

	res, err := r.db.Exec(insertGpsPointsBatch, gpsIDs, bearings, latitudes, longitudes, timestamps)
	if err != nil {
		return 0, err
	}

	inserted, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(inserted), nil
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
