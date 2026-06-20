package gpspoints

import "github.com/ize-302/beacon/backend/internal/common"

type Broadcaster interface {
	Broadcast(CreateGpsPoint)
}

type GpsPointService struct {
	Repository *GpsPointRepository
	Hub        Broadcaster
}

func NewGpsPointService(repository *GpsPointRepository, hub Broadcaster) *GpsPointService {
	return &GpsPointService{Repository: repository, Hub: hub}
}

func (s *GpsPointService) SaveGpsPoint(input *CreateGpsPointRequest) (*common.BaseResponseBody[GpsPointResponse], error) {
	resp := &common.BaseResponseBody[GpsPointResponse]{}
	gpspoint, err := s.Repository.SaveGpsPointRepo(input)
	if err != nil {
		return nil, err
	}
	if s.Hub != nil {
		s.Hub.Broadcast(*input.Body)
	}
	resp.Body.Data = *gpspoint
	resp.Body.Message = "GPS point recorded successfully"
	resp.Body.Status = true
	return resp, nil
}

func (s *GpsPointService) FetchGpsPoints() (*common.BaseResponseBody[[]GpsPointResponse], error) {
	resp := &common.BaseResponseBody[[]GpsPointResponse]{}
	gpspoints, err := s.Repository.FetchGpsPointsRepo()
	if err != nil {
		return nil, err
	}
	list := []GpsPointResponse{}
	if gpspoints != nil {
		list = *gpspoints
	}
	resp.Body.Data = list
	resp.Body.Message = "GPS points fetched successfully"
	resp.Body.Status = true
	return resp, nil
}
