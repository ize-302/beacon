package gps

import "github.com/ize-302/beacon/backend/internal/common"

type GpsService struct {
	Repository *GpsRepository
	EventHub   *EventHub
}

func NewGpsService(repository *GpsRepository, eventHub *EventHub) *GpsService {
	return &GpsService{Repository: repository, EventHub: eventHub}
}

func (s *GpsService) CreateGps(input *CreateGpsRequest) (*common.BaseResponseBody[GpsResponse], error) {
	resp := &common.BaseResponseBody[GpsResponse]{}
	gps, err := s.Repository.CreateGpsRepo(input)
	if err != nil {
		return nil, err
	}
	s.EventHub.Publish(*gps)
	resp.Body.Data = *gps
	resp.Body.Message = "GPS device successfully created"
	resp.Body.Status = true
	return resp, nil
}

func (s *GpsService) FetchGpsDevices() (*common.BaseResponseBody[[]GpsResponse], error) {
	resp := &common.BaseResponseBody[[]GpsResponse]{}
	gpsDevices, err := s.Repository.FetchGpsDevicesRepo()
	if err != nil {
		return nil, err
	}
	list := []GpsResponse{}
	if gpsDevices != nil {
		list = *gpsDevices
	}
	resp.Body.Data = list
	resp.Body.Message = "GPS devices fetched successfully"
	resp.Body.Status = true
	return resp, nil
}

func (s *GpsService) FetchGps(input *GetGpsParams) (*common.BaseResponseBody[GpsResponse], error) {
	resp := &common.BaseResponseBody[GpsResponse]{}
	gps, err := s.Repository.FetchGpsRepo(input)
	if err != nil {
		return nil, err
	}
	resp.Body.Data = *gps
	resp.Body.Message = "GPS device fetched successfully"
	resp.Body.Status = true
	return resp, nil
}

func (s *GpsService) DeleteGps(input *DeleteGpsParams) (*struct{}, error) {
	return nil, s.Repository.DeleteGpsRepo(input)
}

func (s *GpsService) FetchGpsHistory(input *GetGpsHistoryParams) (*common.BaseResponseBody[GpsHistoryResponse], error) {
	resp := &common.BaseResponseBody[GpsHistoryResponse]{}
	history, err := s.Repository.FetchGpsHistoryRepo(input)
	if err != nil {
		return nil, err
	}
	resp.Body.Data = *history
	resp.Body.Message = "GPS history fetched successfully"
	resp.Body.Status = true
	return resp, nil
}
