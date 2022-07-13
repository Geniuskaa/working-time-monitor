package postgres

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"reflect"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/model"
	"testing"
)

func initSqlmock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDb, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	db := sqlx.NewDb(mockDb, "sqlmock")
	return db, sqlMock
}

func TestDb_GetMobileDevices(t *testing.T) {
	db, sqlMock := initSqlmock(t)
	defer db.Close()
	columns := []string{"d_id", "d_name", "d_os"}

	sqlMock.ExpectQuery("SELECT (.+) FROM mobile_devices d").WillReturnRows(
		sqlmock.NewRows(columns).FromCSVString("1,Device 1,IOS\n2,Device 2,ANDROID"))
	d := &Db{Db: db}

	want := []*model.MobileDevice{
		{Id: 1, Name: "Device 1", Os: "IOS"},
		{Id: 2, Name: "Device 2", Os: "ANDROID"},
	}
	wantErr := false

	got, err := d.GetMobileDevicesByOs(context.Background(), "IOS")
	if (err != nil) != wantErr {
		t.Errorf("GetMobileDevicesByOs() error = %v, wantErr %v", err, wantErr)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetMobileDevicesByOs() got = %v, want %v", got, want)
	}
}

func TestDb_GetMobileDevicesByOs_GetIOSDevices(t *testing.T) {
	db, sqlMock := initSqlmock(t)
	defer db.Close()
	columns := []string{"d_id", "d_name", "d_os"}

	sqlMock.ExpectQuery("SELECT (.+) FROM mobile_devices d").WithArgs("IOS").WillReturnRows(
		sqlmock.NewRows(columns).FromCSVString("1,Device 1,IOS"))
	d := &Db{Db: db}

	want := []*model.MobileDevice{
		{Id: 1, Name: "Device 1", Os: "IOS"},
	}
	wantErr := false

	got, err := d.GetMobileDevicesByOs(context.Background(), "IOS")
	if (err != nil) != wantErr {
		t.Errorf("GetMobileDevicesByOs() error = %v, wantErr %v", err, wantErr)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetMobileDevicesByOs() got = %v, want %v", got, want)
	}
}

func TestDb_GetMobileDevicesByOs_GetAndroidDevices(t *testing.T) {
	db, sqlMock := initSqlmock(t)
	defer db.Close()
	columns := []string{"d_id", "d_name", "d_os"}

	sqlMock.ExpectQuery("SELECT (.+) FROM mobile_devices d").WithArgs("IOS").WillReturnRows(
		sqlmock.NewRows(columns).FromCSVString("1,Device 1,ANDROID"))
	d := &Db{Db: db}

	want := []*model.MobileDevice{
		{Id: 1, Name: "Device 1", Os: "ANDROID"},
	}
	wantErr := false

	got, err := d.GetMobileDevicesByOs(context.Background(), "IOS")
	if (err != nil) != wantErr {
		t.Errorf("GetMobileDevicesByOs() error = %v, wantErr %v", err, wantErr)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetMobileDevicesByOs() got = %v, want %v", got, want)
	}
}
