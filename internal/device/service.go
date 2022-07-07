package device

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"time"
)

type Service interface {
	GetMobileDevices(ctx context.Context, os string) ([]*RentingDeviceResponse, error)
	RentDevice(ctx context.Context, deviceId int, userId int) (*RentingDeviceResponse, error)
}

type service struct {
	log  *zap.SugaredLogger
	repo Repository
}

func NewService(log *zap.SugaredLogger, repository Repository) Service {
	return &service{repo: repository, log: log}
}

func (s *service) GetMobileDevices(ctx context.Context, os string) ([]*RentingDeviceResponse, error) {
	devices, err := s.repo.GetMobileDevicesByOs(ctx, os)
	if err != nil {
		return nil, err
	}

	rentingDeviceResponses := make([]*RentingDeviceResponse, 0, len(devices))
	for _, device := range devices {
		rentingDevice, err := s.repo.GetLatestRentingDeviceByDeviceId(ctx, device.Id)
		var free bool
		if err != nil {
			free = true
		} else {
			free = !rentingDevice.UpdatedAt.Valid
		}
		rentingDeviceResponses =
			append(rentingDeviceResponses, &RentingDeviceResponse{Id: device.Id, Name: device.Name, Free: free})
	}

	return rentingDeviceResponses, nil
}

func (s *service) RentDevice(ctx context.Context, deviceId int, userId int) (*RentingDeviceResponse, error) {
	d := RentingDevice{
		UserId:       userId,
		MobileDevice: MobileDevice{Id: deviceId},
		CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	id, err := s.repo.SaveRentingDevice(ctx, &d)
	if err != nil {
		return nil, err
	}

	newDevice, err := s.repo.GetRentingDeviceById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RentingDeviceResponse{Id: newDevice.Id, Name: newDevice.MobileDevice.Name, Free: false}, nil
}
