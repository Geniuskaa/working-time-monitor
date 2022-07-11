package user

import (
	"context"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"mime/multipart"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/auth"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"strconv"
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

func (s *Service) parseXlsxToGetProfiles(file multipart.File, sheetName string) ([]UserProfileDTO, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		s.log.Error(err)
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			s.log.Error(err)
		}
	}()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		s.log.Error(err)
		return nil, err
	}

	users := make([]UserProfileDTO, len(rows)-1)

	for i, row := range rows {
		if i == 0 {
			continue
		}

		id, err := strconv.Atoi(row[0])
		if err != nil {
			s.log.Error(err)
			return nil, err
		}

		j := i - 1
		users[j].Id = id
		users[j].DisplayName = row[1]
		users[j].Employee = row[3]
		users[j].Phone = row[6]
		users[j].Email = row[7]
		if len(row) >= 13 {
			users[j].Devices = row[12]
			if len(row) == 14 {
				users[j].Skills = row[13]
			}
		}

	}
	return users, nil
}

func (s *Service) addSkillToUserByUserUserPrincipal(ctx context.Context, principal *auth.UserPrincipal, skills string) error {
	err := s.repo.AddSkillsToUserProfile(ctx, principal.Username, principal.Email, skills)
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}
