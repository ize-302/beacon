package vehicles

import "github.com/ize-302/beacon/backend/internal/common"

type VehicleService struct {
	Repository *VehicleRepository
	EventHub   *EventHub
}

func NewVehicleService(repository *VehicleRepository, eventHub *EventHub) *VehicleService {
	return &VehicleService{Repository: repository, EventHub: eventHub}
}

func (s *VehicleService) CreateVehicle(input *CreateVehicleRequest) (*common.BaseResponseBody[VehicleResponse], error) {
	resp := &common.BaseResponseBody[VehicleResponse]{}
	vehicle, err := s.Repository.CreateVehicleRepo(input)
	if err != nil {
		return nil, err
	}
	// Announce vehicle creation immediately so new vehicle can start moving
	if s.EventHub != nil {
		s.EventHub.Publish(*vehicle)
	}
	resp.Body.Data = *vehicle
	resp.Body.Message = "Vehicle successfully created"
	resp.Body.Status = true
	return resp, nil
}

func (s *VehicleService) FetchVehicles() (*common.BaseResponseBody[[]VehicleResponse], error) {
	resp := &common.BaseResponseBody[[]VehicleResponse]{}
	vehicles, err := s.Repository.FetchVehiclesRepo()
	if err != nil {
		return nil, err
	}
	resp.Body.Data = vehicles
	resp.Body.Message = "Vehicles fetched successfully"
	resp.Body.Status = true
	return resp, nil
}

func (s *VehicleService) FetchVehicle(input *GetVehicleParams) (*common.BaseResponseBody[VehicleResponse], error) {
	resp := &common.BaseResponseBody[VehicleResponse]{}
	vehicle, err := s.Repository.FetchVehicleRepo(input)
	if err != nil {
		return nil, err
	}
	resp.Body.Data = *vehicle
	resp.Body.Message = "Vehicle successfully fetched"
	resp.Body.Status = true
	return resp, nil
}

func (s *VehicleService) DeleteVehicle(input *DeleteVehicleParams) (*struct{}, error) {
	return nil, s.Repository.DeleteVehicleRepo(input)
}

func (s *VehicleService) FetchVehicleHistory(input *GetVehicleHistoryParams) (*common.BaseResponseBody[VehicleHistoryResponse], error) {
	resp := &common.BaseResponseBody[VehicleHistoryResponse]{}
	history, err := s.Repository.FetchVehicleHistoryRepo(input)
	if err != nil {
		return nil, err
	}
	resp.Body.Data = *history
	resp.Body.Message = "Vehicle history fetched successfully"
	resp.Body.Status = true
	return resp, nil
}
