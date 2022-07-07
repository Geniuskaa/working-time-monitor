package user

import (
	"context"
	"go.uber.org/zap"
)

type Service struct {
	repo *repository
	log  *zap.SugaredLogger
}

func NewService(repo *repository, log *zap.SugaredLogger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) getUsersByEmployeeId(ctx context.Context, id int) ([]UserWithProjectsDTO, error) {
	users, err := s.repo.getUsersByEmplId(ctx, id)
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
	emplList, err := s.repo.getEmplList(ctx)
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
