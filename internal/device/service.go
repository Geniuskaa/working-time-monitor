package device

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/model"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"time"
)

type Service interface {
	GetMobileDevices(ctx context.Context, os string) ([]*RentingDeviceResponse, error)
	RentDevice(ctx context.Context, deviceId int, userId int) (*RentingDeviceResponse, error)
	ReturnDevice(ctx context.Context, deviceId int, userId int) (*RentingDeviceResponse, error)
}

type service struct {
	log  *zap.SugaredLogger
	repo postgres.DeviceRepo
}

func NewService(log *zap.SugaredLogger, repository postgres.DeviceRepo) Service {
	return &service{repo: repository, log: log}
}

func (s *service) GetMobileDevices(ctx context.Context, os string) ([]*RentingDeviceResponse, error) {
	var devices []*model.MobileDevice
	var err error
	if os == "" {
		devices, err = s.repo.GetMobileDevices(ctx)
	} else {
		devices, err = s.repo.GetMobileDevicesByOs(ctx, os)
	}
	if err != nil {
		return nil, err
	}

	rentingDeviceResponses := make([]*RentingDeviceResponse, 0, len(devices))
	for _, device := range devices {
		owner := s.getMobileDeviceOwner(ctx, device.Id)
		rentingDeviceResponses = append(
			rentingDeviceResponses,
			&RentingDeviceResponse{
				Id:          device.Id,
				Name:        device.Name,
				DisplayName: owner.DisplayName})
	}
	return rentingDeviceResponses, nil
}

func (s *service) RentDevice(ctx context.Context, deviceId int, userId int) (*RentingDeviceResponse, error) {
	owner := s.getMobileDeviceOwner(ctx, deviceId)
	if !owner.Free {
		return nil, errors.New("mobile device already rented")
	}
	d := model.RentingDevice{
		User:         model.RentingDeviceUser{Id: userId},
		MobileDevice: model.MobileDevice{Id: deviceId},
		CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	id, err := s.repo.SaveRentingDevice(ctx, &d)
	if err != nil {
		return nil, err
	}

	newRentingDevice, err := s.repo.GetRentingDeviceById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RentingDeviceResponse{
		Id:          deviceId,
		Name:        newRentingDevice.MobileDevice.Name,
		DisplayName: newRentingDevice.User.DisplayName,
	}, nil
}

func (s *service) ReturnDevice(ctx context.Context, deviceId int, userId int) (*RentingDeviceResponse, error) {
	latestRentingDevice, err := s.repo.GetLatestRentingDeviceByDeviceId(ctx, deviceId)
	if err != nil {
		return nil, err
	}

	if latestRentingDevice.User.Id != userId {
		return nil, errors.New("the user who rented the device and who is trying to return are not the same")
	}
	return nil, nil
}

type mobileDeviceOwner struct {
	DisplayName string
	Free        bool
}

func (s *service) getMobileDeviceOwner(ctx context.Context, deviceId int) *mobileDeviceOwner {
	rentingDevice, err := s.repo.GetLatestRentingDeviceByDeviceId(ctx, deviceId)
	var free bool
	if err != nil {
		free = true
	} else {
		free = !rentingDevice.UpdatedAt.Valid
	}
	var displayName string
	if !free {
		displayName = rentingDevice.User.DisplayName
	}
	return &mobileDeviceOwner{
		DisplayName: displayName,
		Free:        free,
	}
}
