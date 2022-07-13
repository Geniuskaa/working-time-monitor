package device

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"reflect"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/mocks"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/model"
	"testing"
	"time"
)

var (
	mobileDevices = []*model.MobileDevice{
		{Id: 1, Name: "Device1", Os: "IOS"},
		{Id: 2, Name: "Device2", Os: "ANDROID"},
	}

	rentingDevices = []*model.RentingDevice{
		{
			Id:           1,
			MobileDevice: *mobileDevices[0],
			User:         model.RentingDeviceUser{Id: 1, Username: "TestUser1", DisplayName: "Test User1"},
			CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
		{
			Id:           2,
			MobileDevice: *mobileDevices[1],
			User:         model.RentingDeviceUser{Id: 1, Username: "TestUser2", DisplayName: "Test User2"},
			CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}
)

func initMockLogger(t *testing.T) *zap.SugaredLogger {
	return zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		if e.Level == zap.ErrorLevel {
			t.Fatal("Error should never happen!")
		}
		return nil
	}))).Sugar()
}

func Test_service_GetMobileDevices(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockDeviceRepo(ctrl)
	mockRepo.EXPECT().GetMobileDevices(gomock.Any()).Return(mobileDevices, nil).AnyTimes()
	mockRepo.EXPECT().GetMobileDevicesByOs(gomock.Any(), "IOS").Return(mobileDevices[1:], nil).AnyTimes()
	mockRepo.EXPECT().GetMobileDevicesByOs(gomock.Any(), "ANDROID").Return(mobileDevices[:1], nil).AnyTimes()
	mockRepo.EXPECT().GetLatestRentingDeviceByDeviceId(gomock.Any(), 1).Return(rentingDevices[0], nil).AnyTimes()
	mockRepo.EXPECT().GetLatestRentingDeviceByDeviceId(gomock.Any(), 2).Return(rentingDevices[1], nil).AnyTimes()

	s := service{log: initMockLogger(t), repo: mockRepo}

	type args struct {
		ctx context.Context
		os  string
	}
	tests := []struct {
		name    string
		args    args
		want    []*RentingDeviceResponse
		wantErr bool
	}{
		{
			name: "Get all devices",
			args: args{context.Background(), ""},
			want: []*RentingDeviceResponse{
				{
					Id:          1,
					Name:        "Device1",
					DisplayName: "Test User1",
				},
				{
					Id:          2,
					Name:        "Device2",
					DisplayName: "",
				},
			},
			wantErr: false,
		},
		{
			name: "Get IOS devices",
			args: args{context.Background(), "IOS"},
			want: []*RentingDeviceResponse{
				{
					Id:          2,
					Name:        "Device2",
					DisplayName: "",
				},
			},
			wantErr: false,
		},
		{
			name: "Get Android devices",
			args: args{context.Background(), "ANDROID"},
			want: []*RentingDeviceResponse{
				{
					Id:          1,
					Name:        "Device1",
					DisplayName: "Test User1",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetMobileDevices(tt.args.ctx, tt.args.os)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMobileDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMobileDevices() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_RentDevice(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockDeviceRepo(ctrl)

	mockRepo.EXPECT().GetLatestRentingDeviceByDeviceId(gomock.Any(), 1).Return(rentingDevices[0], nil).AnyTimes()
	mockRepo.EXPECT().GetLatestRentingDeviceByDeviceId(gomock.Any(), 2).Return(rentingDevices[1], nil).AnyTimes()
	mockRepo.EXPECT().SaveRentingDevice(gomock.Any(), gomock.Any()).Return(3, nil).AnyTimes()
	mockRepo.EXPECT().GetRentingDeviceById(gomock.Any(), 3).Return(
		&model.RentingDevice{
			Id:           3,
			MobileDevice: model.MobileDevice{Id: 2},
			User:         model.RentingDeviceUser{Id: 1, Username: "TestUser", DisplayName: "Test User1"},
			CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true}},
		nil,
	).AnyTimes()
	mockRepo.EXPECT().GetMobileDeviceById(gomock.Any(), 2).Return(mobileDevices[1], nil).AnyTimes()

	s := service{log: initMockLogger(t), repo: mockRepo}

	type args struct {
		ctx      context.Context
		deviceId int
		userId   int
	}
	tests := []struct {
		name    string
		args    args
		want    *RentingDeviceResponse
		wantErr bool
	}{
		{
			name: "Successfully rent device",
			args: args{ctx: context.Background(), deviceId: 2, userId: 1},
			want: &RentingDeviceResponse{
				Id:          2,
				Name:        mobileDevices[1].Name,
				DisplayName: "Test User1",
			},
			wantErr: false,
		},
		{
			name:    "Try to rent device which is already rented",
			args:    args{ctx: context.Background(), deviceId: 1, userId: 1},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.RentDevice(tt.args.ctx, tt.args.deviceId, tt.args.userId)
			if (err != nil) != tt.wantErr {
				t.Errorf("RentDevice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RentDevice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_ReturnDevice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockDeviceRepo(ctrl)

	mockRepo.EXPECT().GetLatestRentingDeviceByDeviceId(gomock.Any(), 1).Return(rentingDevices[0], nil).AnyTimes()
	mockRepo.EXPECT().GetLatestRentingDeviceByDeviceId(gomock.Any(), 2).Return(rentingDevices[1], nil).AnyTimes()
	mockRepo.EXPECT().GetMobileDeviceById(gomock.Any(), 1).Return(mobileDevices[0], nil).AnyTimes()
	mockRepo.EXPECT().GetMobileDeviceById(gomock.Any(), 2).Return(mobileDevices[1], nil).AnyTimes()

	mockRepo.EXPECT().UpdateRentingDevice(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	s := &service{log: initMockLogger(t), repo: mockRepo}

	type args struct {
		ctx      context.Context
		deviceId int
		userId   int
	}
	tests := []struct {
		name    string
		args    args
		want    *RentingDeviceResponse
		wantErr bool
	}{
		{
			name:    "User who rented the device and who is trying to return are not the same",
			args:    args{ctx: context.Background(), deviceId: 1, userId: 2},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Device with given device id is not rented",
			args:    args{ctx: context.Background(), deviceId: 2, userId: 1},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Successfully return device",
			args: args{ctx: context.Background(), deviceId: 1, userId: 1},
			want: &RentingDeviceResponse{
				Id:          1,
				Name:        mobileDevices[0].Name,
				DisplayName: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.ReturnDevice(tt.args.ctx, tt.args.deviceId, tt.args.userId)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReturnDevice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReturnDevice() got = %v, want %v", got, tt.want)
			}
		})
	}
}
