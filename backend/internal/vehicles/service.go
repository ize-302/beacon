package vehicles

import "github.com/ize-302/beacon/backend/internal/common"

type VehicleService struct {
	Repository *VehicleRepository
}

func NewVehicleService(repository *VehicleRepository) *VehicleService {
	return &VehicleService{Repository: repository}
}

func ToVehicleResponse(vehicle *Vehicle) VehicleResponse {
	return VehicleResponse{
		ID:          vehicle.ID,
		PlateNumber: vehicle.PlateNumber,
		VehicleType: vehicle.VehicleType,
		CreatedAt:   vehicle.CreatedAt,
	}
}

func (s *VehicleService) CreateVehicle(input *CreateVehicleRequest) (*common.BaseResponseBody[VehicleResponse], error) {
	resp := &common.BaseResponseBody[VehicleResponse]{}
	vehicle, err := s.Repository.CreateVehicleRepo(input)
	if err != nil {
		return nil, err
	}
	resp.Body.Data.ID = vehicle.ID
	resp.Body.Data.PlateNumber = vehicle.PlateNumber
	resp.Body.Data.VehicleType = vehicle.VehicleType
	resp.Body.Data.CreatedAt = vehicle.CreatedAt
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
	vehiclesList := []VehicleResponse{}
	for _, vehicle := range *vehicles {
		vehiclesList = append(vehiclesList, ToVehicleResponse(&vehicle))
	}
	resp.Body.Data = vehiclesList
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
	resp.Body.Data = ToVehicleResponse(vehicle)
	resp.Body.Message = "Vehicle successfully fetched"
	resp.Body.Status = true
	return resp, nil
}

func (s *VehicleService) DeleteVehicle(input *DeleteVehicleParams) (*struct{}, error) {
	return nil, s.Repository.DeleteVehicleRepo(input)
}
