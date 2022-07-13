package user

import (
	"context"
	"fmt"
	"github.com/xuri/excelize/v2"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"mime/multipart"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/auth"
	"scb-mobile/scb-monitor/scb-monitor-backend/go-app/internal/postgres"
	"strings"
)

type Service struct {
	repo *postgres.Db
	log  *zap.SugaredLogger
}

func NewService(repo *postgres.Db, log *zap.SugaredLogger) *Service {
	return &Service{repo: repo, log: log}
}

var ErrNotAllProfilesWereAdded error

func (s *Service) getUsersByEmployeeId(ctx context.Context, id int) ([]UserWithProjectsDTO, error) {
	tr := otel.Tracer("component-getEmployeeList")
	_, span := tr.Start(ctx, "service-getEmployeeList")
	//span.SetAttributes(attribute.Key("testset").String("value"))
	defer span.End()

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
	tr := otel.Tracer("component-getEmployeeList")
	_, span := tr.Start(ctx, "service-getEmployeeList")
	//span.SetAttributes(attribute.Key("testset").String("value"))
	defer span.End()

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

func (s *Service) getUser(ctx context.Context, userId int) (UserDTO, error) {
	tr := otel.Tracer("component-getUser")
	_, span := tr.Start(ctx, "service-getUser")
	//span.SetAttributes(attribute.Key("testset").String("value"))
	defer span.End()

	user, empl, err := s.repo.GetUserByUserId(ctx, userId)
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

func (s *Service) parseXlsxToGetProfiles(ctx context.Context, file multipart.File, sheetName string) error {
	tr := otel.Tracer("component-parseXlsxToGetProfiles")
	_, span := tr.Start(ctx, "service-parseXlsxToGetProfiles")
	//span.SetAttributes(attribute.Key("testset").String("value"))
	defer span.End()

	f, err := excelize.OpenReader(file)
	if err != nil {
		s.log.Error(err)
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			s.log.Error(err)
		}
	}()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		s.log.Error(err)
		return err
	}

	users := make([]postgres.UserProfileFromExcel, len(rows)-1)

	for i, row := range rows {
		if i == 0 { // Пропускаем строку с наименованиями столбцов
			continue
		}

		if len(row) < 8 { // Данных меньше минимума необходимого для заполнения профиля, поэтому пропускаем эту строку
			continue
		}

		j := i - 1
		users[j].DisplayName = row[1]
		users[j].Employee = fmt.Sprintf(row[3] + " " + row[4])
		users[j].Phone = row[6]
		users[j].Email = row[7]

		if len(row) >= 11 { // поиск личных устройств

			startOfDevicesid := 8 // 4 устройства могут прийти из Excel файла

			users[j].Devices = make([]postgres.Device, 4)

			for i := 0; i < 4; i++ {

				if len(row) == 11 && i == 3 { //Если у нас данных ровно на 11 колонок - мы скипнем обращение к 12-му элементу
					continue
				}

				isPersonalDevice := strings.ContainsAny(row[startOfDevicesid+i], "(K3")
				if isPersonalDevice {
					isMonitor := strings.HasPrefix(strings.ToLower(row[startOfDevicesid+i]), "монитор")
					if isMonitor {
						users[j].Devices[i] = postgres.Device{
							Name: row[startOfDevicesid+i],
							Type: "Монитор",
						}
						continue
					}

					users[j].Devices[i] = postgres.Device{
						Name: row[startOfDevicesid+i],
						Type: "ПК",
					}

				}

			}
		}

		if len(row) >= 13 {
			users[j].MobileDevices = strings.Split(row[12], ",")
			if len(row) == 14 {
				users[j].Skills = row[13]
			}
		}

	}

	countOfInserts, err := s.repo.PutProfilesToDB(ctx, users)
	if err != nil {
		s.log.Error(err)
		return err
	}

	if countOfInserts != len(rows)-1 {
		s.log.Error(ErrNotAllProfilesWereAdded, "Не все профили были добавлены в БД")
		return ErrNotAllProfilesWereAdded
	}

	return nil
}

func (s *Service) addSkillToUser(ctx context.Context, userPrincipal *auth.UserPrincipal, skills string) error {
	tr := otel.Tracer("component-addSkillToUser")
	_, span := tr.Start(ctx, "service-addSkillToUser")
	//span.SetAttributes(attribute.Key("testset").String("value"))
	defer span.End()

	err := s.repo.AddSkillsToUserProfile(ctx, userPrincipal.Username, userPrincipal.Email, skills)
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}
