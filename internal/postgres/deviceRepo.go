package postgres

import (
	"context"
	"errors"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/model"
)

type DeviceRepo interface {
	GetMobileDeviceById(ctx context.Context, id int) (*model.MobileDevice, error)
	GetMobileDevices(ctx context.Context) ([]*model.MobileDevice, error)
	GetMobileDevicesByOs(ctx context.Context, os string) ([]*model.MobileDevice, error)
	GetLatestRentingDeviceByDeviceId(ctx context.Context, deviceId int) (*model.RentingDevice, error)
	GetRentingDeviceById(ctx context.Context, id int) (*model.RentingDevice, error)
	SaveRentingDevice(ctx context.Context, rentingDevice *model.RentingDevice) (int, error)
	UpdateRentingDevice(ctx context.Context, id int, device *model.RentingDevice) error
}

func (d *Db) GetMobileDevices(ctx context.Context) ([]*model.MobileDevice, error) {
	q := "SELECT d.id, d.name, d.os  FROM mobile_devices d"
	rows, err := d.Db.QueryContext(ctx, q)
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
	q := "SELECT d.id, d.name, d.os  FROM mobile_devices d WHERE d.os = $1"
	rows, err := d.Db.QueryContext(ctx, q, os)
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
	q := `SELECT d.id, d.created_at, d.updated_at, u.id, u.username, u.display_name 
FROM renting_devices d INNER JOIN users u ON d.user_id = u.id 
WHERE d.mobile_device_id = $1 
ORDER BY d.created_at DESC LIMIT 1`
	row := d.Db.QueryRowContext(ctx, q, deviceId)
	device := model.RentingDevice{MobileDevice: model.MobileDevice{Id: deviceId}}
	err := row.Scan(&device.Id, &device.CreatedAt, &device.UpdatedAt, &device.User.Id, &device.User.Username, &device.User.DisplayName)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (d *Db) GetRentingDeviceById(ctx context.Context, id int) (*model.RentingDevice, error) {
	q := `SELECT d.id, d.mobile_device_id, d.created_at, d.updated_at, u.id, u.username, u.display_name 
FROM renting_devices d INNER JOIN users u on d.user_id = u.id 
WHERE d.id = $1`
	row := d.Db.QueryRowContext(ctx, q, id)
	device := model.RentingDevice{}
	err := row.Scan(&device.Id, &device.MobileDevice.Id, &device.CreatedAt, &device.UpdatedAt, &device.User.Id,
		&device.User.Username, &device.User.DisplayName)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (d *Db) SaveRentingDevice(ctx context.Context, rentingDevice *model.RentingDevice) (int, error) {
	q := `INSERT INTO renting_devices(mobile_device_id, user_id, created_at) VALUES ($1, $2, $3) RETURNING id`
	id := 0
	err := d.Db.QueryRowContext(ctx, q, rentingDevice.MobileDevice.Id, rentingDevice.User.Id, rentingDevice.CreatedAt).Scan(&id)
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func (d *Db) UpdateRentingDevice(ctx context.Context, id int, device *model.RentingDevice) error {
	q := `UPDATE renting_devices SET mobile_device_id = $2, user_id = $3, updated_at = $4 WHERE id = $1`
	result, err := d.Db.Exec(q, id, device.MobileDevice.Id, device.User.Id, device.UpdatedAt)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if rowsAffected != 1 {
		return errors.New("update renting device error")
	}
	return nil
}

func (d *Db) GetMobileDeviceById(ctx context.Context, id int) (*model.MobileDevice, error) {
	q := `SELECT id, name, os FROM mobile_devices WHERE id = $1`
	row := d.Db.QueryRowContext(ctx, q, id)
	device := model.MobileDevice{}
	err := row.Scan(&device.Id, &device.Name, &device.Os)
	if err != nil {
		return nil, err
	}
	return &device, nil
}
