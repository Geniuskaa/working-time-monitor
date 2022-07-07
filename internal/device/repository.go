package device

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetMobileDevices(ctx context.Context) ([]*MobileDevice, error)
	GetMobileDevicesByOs(ctx context.Context, os string) ([]*MobileDevice, error)
	GetLatestRentingDeviceByDeviceId(ctx context.Context, deviceId int) (*RentingDevice, error)
	GetRentingDeviceById(ctx context.Context, id int) (*RentingDevice, error)
	SaveRentingDevice(ctx context.Context, rentingDevice *RentingDevice) (int, error)
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

type repository struct {
	db *sqlx.DB
}

func (r *repository) GetMobileDevices(ctx context.Context) ([]*MobileDevice, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT d.id, d.name, d.os  FROM mobile_devices d")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*MobileDevice, 0)
	for rows.Next() {
		d := MobileDevice{}
		err := rows.Scan(&d.Id, &d.Name, &d.Os)
		if err != nil {
			return nil, err
		}
		result = append(result, &d)
	}
	return result, nil
}

func (r *repository) GetMobileDevicesByOs(ctx context.Context, os string) ([]*MobileDevice, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT d.id, d.name, d.os  FROM mobile_devices d WHERE d.os = ?", os)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*MobileDevice, 0)
	for rows.Next() {
		d := MobileDevice{}
		err := rows.Scan(&d.Id, &d.Name, &d.Os)
		if err != nil {
			return nil, err
		}
		result = append(result, &d)
	}
	return result, nil
}

func (r *repository) GetLatestRentingDeviceByDeviceId(ctx context.Context, deviceId int) (*RentingDevice, error) {
	row := r.db.QueryRowContext(ctx, "SELECT d.id, d.user_id, d.created_at, d.updated_at FROM renting_devices d "+
		"WHERE d.mobile_device_id = ? ORDER BY d.created_at DESC LIMIT 1", deviceId)
	d := RentingDevice{}
	err := row.Scan(&d.Id, &d.UserId, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *repository) GetRentingDeviceById(ctx context.Context, id int) (*RentingDevice, error) {
	row := r.db.QueryRowContext(ctx, "SELECT d.id, d.mobile_device_id, d.user_id, d.created_at, d.updated_at "+
		"FROM renting_devices d WHERE id = ?", id)
	d := RentingDevice{}
	err := row.Scan(&d.Id, &d.MobileDevice.Id, &d.UserId, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *repository) SaveRentingDevice(ctx context.Context, rentingDevice *RentingDevice) (int, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO renting_devices(mobile_device_id, user_id) VALUES (?, ?)",
		rentingDevice.MobileDevice.Id, rentingDevice.UserId)
	if err != nil {
		return -1, err
	}
	lastInsertId, err := result.LastInsertId()
	return int(lastInsertId), nil
}
