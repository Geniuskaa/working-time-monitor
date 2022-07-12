package user

import (
	"context"
	"go.uber.org/zap"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
)

type Service struct {
	repo *postgres.Db
	log  *zap.SugaredLogger
}

func NewService(repo *postgres.Db, log *zap.SugaredLogger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) getUsersByEmployeeId(ctx context.Context, id int) ([]UserWithProjectsDTO, error) {
	users, err := s.repo.GetUsersByEmplId(ctx, id)
	if err != nil {
		s.log.Error(err)
		return nil, err
	}

	usersDto := make([]UserWithProjectsDTO, len(users))
	for i, user := range users {

		usersDto[i] = UserWithProjectsDTO{
			Id:          user.Id,
			DisplayName: user.DisplayName,
			Projects:    user.Projects,
		}
	}

	return usersDto, nil
}

func (s *Service) getEmployeeList(ctx context.Context) ([]EmpolyeeDTO, error) {
	emplList, err := s.repo.GetEmplList(ctx)
	if err != nil {
		s.log.Error(err)
		return nil, err
	}

	emplListDto := make([]EmpolyeeDTO, len(emplList))

	for i, empl := range emplList {
		emplListDto[i] = EmpolyeeDTO{
			Id:   empl.Id,
			Name: empl.Name,
		}
	}

	return emplListDto, nil
}

func (s *Service) getUserByUserId(ctx context.Context, id int) (UserDTO, error) {
	user, empl, err := s.repo.GetUserByUserId(ctx, id)
	if err != nil {
		s.log.Error(err)
		return UserDTO{}, err
	}

	userDto := UserDTO{
		Id:          user.Id,
		DisplayName: user.DisplayName,
		Employee:    empl.Name,
		Email:       user.Email,
		Phone:       user.Phone,
		Birthday:    user.Birthday,
		Skills:      user.Skills,
	}

	return userDto, nil
}
