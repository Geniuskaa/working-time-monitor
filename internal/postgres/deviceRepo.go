package postgres

import (
	"context"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/model"
)

type DeviceRepo interface {
	GetMobileDevices(ctx context.Context) ([]*model.MobileDevice, error)
	GetMobileDevicesByOs(ctx context.Context, os string) ([]*model.MobileDevice, error)
	GetLatestRentingDeviceByDeviceId(ctx context.Context, deviceId int) (*model.RentingDevice, error)
	GetRentingDeviceById(ctx context.Context, id int) (*model.RentingDevice, error)
	SaveRentingDevice(ctx context.Context, rentingDevice *model.RentingDevice) (int, error)
}

func (d *Db) GetMobileDevices(ctx context.Context) ([]*model.MobileDevice, error) {
	rows, err := d.Db.QueryContext(ctx, "SELECT d.id, d.name, d.os  FROM mobile_devices d")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*model.MobileDevice, 0)
	for rows.Next() {
		d := model.MobileDevice{}
		err := rows.Scan(&d.Id, &d.Name, &d.Os)
		if err != nil {
			return nil, err
		}
		result = append(result, &d)
	}
	return result, nil
}

func (d *Db) GetMobileDevicesByOs(ctx context.Context, os string) ([]*model.MobileDevice, error) {
	rows, err := d.Db.QueryContext(ctx, "SELECT d.id, d.name, d.os  FROM mobile_devices d WHERE d.os = $1", os)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*model.MobileDevice, 0)
	for rows.Next() {
		d := model.MobileDevice{}
		err := rows.Scan(&d.Id, &d.Name, &d.Os)
		if err != nil {
			return nil, err
		}
		result = append(result, &d)
	}
	return result, nil
}

func (d *Db) GetLatestRentingDeviceByDeviceId(ctx context.Context, deviceId int) (*model.RentingDevice, error) {
	row := d.Db.QueryRowContext(ctx, "SELECT d.id, d.user_id, d.created_at, d.updated_at FROM renting_devices d "+
		"WHERE d.mobile_device_id = $1 ORDER BY d.created_at DESC LIMIT 1", deviceId)
	device := model.RentingDevice{}
	err := row.Scan(&device.Id, &device.UserId, &device.CreatedAt, &device.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (d *Db) GetRentingDeviceById(ctx context.Context, id int) (*model.RentingDevice, error) {
	row := d.Db.QueryRowContext(ctx, "SELECT d.id, d.mobile_device_id, d.user_id, d.created_at, d.updated_at "+
		"FROM renting_devices d WHERE id = $1", id)
	device := model.RentingDevice{}
	err := row.Scan(&device.Id, &device.MobileDevice.Id, &device.UserId, &device.CreatedAt, &device.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (d *Db) SaveRentingDevice(ctx context.Context, rentingDevice *model.RentingDevice) (int, error) {
	result, err := d.Db.ExecContext(ctx, "INSERT INTO renting_devices(mobile_device_id, user_id) VALUES (?, ?)",
		rentingDevice.MobileDevice.Id, rentingDevice.UserId)
	if err != nil {
		return -1, err
	}
	lastInsertId, err := result.LastInsertId()
	return int(lastInsertId), nil
}
