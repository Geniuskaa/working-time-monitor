package user

import (
	"context"
	"database/sql"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"reflect"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/auth"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/mocks"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"testing"
)

func initMockLogger(t *testing.T) *zap.SugaredLogger {
	return zaptest.NewLogger(t, zaptest.Level(zap.ErrorLevel)).Sugar()
}

var (
	userWithProjects = []*postgres.UserWithProjects{
		{1, "Глухов Амир Владиславович", "Халвёнок"},
		{2, "Архипов Алексей Вячаславович", "Халвёнок"},
	}

	employees = []*postgres.Employee{
		{1, "Java-developer"},
		{2, "Golang-developer"},
	}
)

func TestService_getUsersByEmployeeId(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockUserRepo(ctrl)
	mockRepo.EXPECT().GetUsersByEmplId(context.Background(), 1).Return(userWithProjects, nil).AnyTimes()
	mockRepo.EXPECT().GetUsersByEmplId(context.Background(), 1000).Return([]*postgres.UserWithProjects{}, nil).AnyTimes()

	s := &Service{
		repo: mockRepo,
		log:  initMockLogger(t),
	}

	type fields struct {
		repo postgres.UserRepo
		log  *zap.SugaredLogger
	}
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []UserWithProjectsDTO
		wantErr bool
	}{
		{name: "Успешный запрос с существующим Id", fields: fields{
			repo: s.repo,
			log:  s.log,
		}, args: args{
			ctx: context.Background(),
			id:  1,
		}, want: []UserWithProjectsDTO{{1, "Глухов Амир Владиславович", "Халвёнок"},
			{2, "Архипов Алексей Вячаславович", "Халвёнок"}}, wantErr: false},
		{name: "Запрос с несуществующим Id", fields: fields{
			repo: s.repo,
			log:  s.log,
		}, args: args{
			ctx: context.Background(),
			id:  1000,
		}, want: []UserWithProjectsDTO{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := s.getUsersByEmployeeId(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUsersByEmployeeId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUsersByEmployeeId() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_getEmployeeList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockUserRepo(ctrl)
	mockRepo.EXPECT().GetEmplList(gomock.Any()).Return(employees, nil).AnyTimes()

	s := &Service{
		repo: mockRepo,
		log:  initMockLogger(t),
	}

	type fields struct {
		repo postgres.UserRepo
		log  *zap.SugaredLogger
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []EmpolyeeDTO
		wantErr bool
	}{
		{name: "Успешный запрос", fields: fields{
			repo: s.repo,
			log:  s.log,
		}, args: args{
			ctx: context.Background(),
		}, want: []EmpolyeeDTO{{1, "Java-developer"},
			{2, "Golang-developer"}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := s.getEmployeeList(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getEmployeeList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEmployeeList() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_getUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockUserRepo(ctrl)
	mockRepo.EXPECT().GetUser(context.Background(), 1).Return(&postgres.User{
		Id:          1,
		Username:    "egorchik",
		DisplayName: "Fazluev Egor Olegovich",
		EmployeeId:  4,
		Email:       "egorka@mail.ru",
		Phone:       "+7956457457",
		Skills:      "PostgreSQL, MongoDB",
	}, &postgres.Employee{
		Id:   4,
		Name: "Data engineer",
	}, nil).AnyTimes()
	mockRepo.EXPECT().GetUser(context.Background(), 1000).Return(nil, nil, sql.ErrNoRows).AnyTimes()

	s := &Service{
		repo: mockRepo,
		log:  initMockLogger(t),
	}

	type fields struct {
		repo postgres.UserRepo
		log  *zap.SugaredLogger
	}
	type args struct {
		ctx    context.Context
		userId int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    UserDTO
		wantErr bool
	}{
		{name: "Успешный запрос с существующим Id", fields: fields{
			repo: s.repo,
			log:  s.log,
		}, args: args{
			ctx:    context.Background(),
			userId: 1,
		}, want: UserDTO{
			Id:          1,
			DisplayName: "Fazluev Egor Olegovich",
			Employee:    "Data engineer",
			Email:       "egorka@mail.ru",
			Phone:       "+7956457457",
			Skills:      "PostgreSQL, MongoDB",
		}, wantErr: false},
		{name: "Запрос с несуществующим Id", fields: fields{
			repo: s.repo,
			log:  s.log,
		}, args: args{
			ctx:    context.Background(),
			userId: 1000,
		}, want: UserDTO{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.getUser(tt.args.ctx, tt.args.userId)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_addSkillToUserByUserUserPrincipal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockUserRepo(ctrl)
	mockRepo.EXPECT().AddSkillsToUserProfile(context.Background(), "testUser", "testik@mail.ru", "Java, Redis").Return(nil).AnyTimes()

	s := &Service{
		repo: mockRepo,
		log:  initMockLogger(t),
	}

	type fields struct {
		repo postgres.UserRepo
		log  *zap.SugaredLogger
	}
	type args struct {
		ctx       context.Context
		principal *auth.UserPrincipal
		skills    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "Успешный запрос", fields: fields{
			repo: s.repo,
			log:  s.log,
		}, args: args{
			ctx: context.Background(),
			principal: &auth.UserPrincipal{
				Username: "testUser",
				Email:    "testik@mail.ru",
			},
			skills: "Java, Redis",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := s.addSkillToUser(tt.args.ctx, tt.args.principal, tt.args.skills); (err != nil) != tt.wantErr {
				t.Errorf("addSkillToUserByUserUserPrincipal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
