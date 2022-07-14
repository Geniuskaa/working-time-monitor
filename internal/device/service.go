package device

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"net/http"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/apperror"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/model"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"time"
)

//go:generate mockgen -destination=../mocks/device_service.go -package=mocks . Service

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

var (
	errDeviceAlreadyRented      = apperror.NewAppError(nil, "Device already rented", http.StatusBadRequest)
	errDeviceIsNotRented        = apperror.NewAppError(nil, "Device is not rented", http.StatusBadRequest)
	errDeviceRenterIsNotTheSame = apperror.NewAppError(nil, "The user who rented the device and who is trying to return are not the same", http.StatusBadRequest)
)

func (s *service) GetMobileDevices(ctx context.Context, os string) ([]*RentingDeviceResponse, error) {
	tr := otel.Tracer("getMobileDevices")
	ctx, span := tr.Start(ctx, "service-GetMobileDevices")
	defer span.End()
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
		displayName := s.getRentingDeviceOwnerDisplayName(ctx, device.Id)
		rentingDeviceResponses = append(
			rentingDeviceResponses,
			&RentingDeviceResponse{
				Id:          device.Id,
				Name:        device.Name,
				DisplayName: displayName})
	}
	return rentingDeviceResponses, nil
}

func (s *service) RentDevice(ctx context.Context, deviceId int, userId int) (*RentingDeviceResponse, error) {
	tr := otel.Tracer("RentDevice")
	ctx, span := tr.Start(ctx, "service-RentDevice")
	defer span.End()
	owner := s.getRentingDeviceOwnerDisplayName(ctx, deviceId)
	if owner != "" {
		return nil, errDeviceAlreadyRented
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

	rentingDevice, err := s.repo.GetRentingDeviceById(ctx, id)
	if err != nil {
		return nil, err
	}
	mobileDevice, err := s.repo.GetMobileDeviceById(ctx, rentingDevice.MobileDevice.Id)
	if err != nil {
		return nil, err
	}
	return &RentingDeviceResponse{
		Id:          deviceId,
		Name:        mobileDevice.Name,
		DisplayName: rentingDevice.User.DisplayName,
	}, nil
}

func (s *service) ReturnDevice(ctx context.Context, deviceId int, userId int) (*RentingDeviceResponse, error) {
	tr := otel.Tracer("ReturnDevice")
	ctx, span := tr.Start(ctx, "service-ReturnDevice")
	defer span.End()
	latestRentingDevice, err := s.repo.GetLatestRentingDeviceByDeviceId(ctx, deviceId)
	if err != nil {
		return nil, err
	}

	if latestRentingDevice.User.Id != userId {
		return nil, errDeviceRenterIsNotTheSame
	}
	if latestRentingDevice.UpdatedAt.Valid {
		return nil, errDeviceIsNotRented
	}
	latestRentingDevice.UpdatedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
	err = s.repo.UpdateRentingDevice(ctx, latestRentingDevice.Id, latestRentingDevice)
	if err != nil {
		return nil, err
	}
	mobileDevice, err := s.repo.GetMobileDeviceById(ctx, latestRentingDevice.MobileDevice.Id)
	if err != nil {
		return nil, err
	}
	return &RentingDeviceResponse{
		Id:   deviceId,
		Name: mobileDevice.Name,
	}, nil
}

func (s *service) getRentingDeviceOwnerDisplayName(ctx context.Context, deviceId int) string {
	rentingDevice, err := s.repo.GetLatestRentingDeviceByDeviceId(ctx, deviceId)
	if err != nil {
		return ""
	}
	if !rentingDevice.UpdatedAt.Valid {
		return rentingDevice.User.DisplayName
	}
	return ""
}
