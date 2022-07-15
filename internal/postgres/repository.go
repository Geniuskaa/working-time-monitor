package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel"
	"strings"
)

//go:generate mockgen -destination=../mocks/user_repo.go -package=mocks . UserRepo

type UserRepo interface {
	GetUsersByEmplId(ctx context.Context, id int) ([]*UserWithProjects, error)
	GetEmplList(ctx context.Context) ([]*Employee, error)
	GetUser(ctx context.Context, userId int) (*User, *Employee, error)
	GetUserPrincipalByUsername(ctx context.Context, username string) (*UserPrincipal, error)
	AddSkillsToUserProfile(ctx context.Context, username string, email string, skills string) error
	PutProfilesToDB(ctx context.Context, users []UserProfileFromExcel) (int, error)
	GetUsersProfiles(ctx context.Context) ([]*UserProfile, error)
}

func (d *Db) GetUserPrincipalByUsername(ctx context.Context, username string) (*UserPrincipal, error) {
	principal := UserPrincipal{}
	row := d.Db.QueryRowContext(ctx, `SELECT u.id, u.username, u.email FROM users u WHERE u.username = $1`, username)
	err := row.Scan(&principal.Id, &principal.Username, &principal.Email)
	if err != nil {
		return nil, err
	}
	return &principal, nil
}

func (d *Db) GetUsersByEmplId(ctx context.Context, id int) ([]*UserWithProjects, error) {
	tr := otel.Tracer("repo-GetUsersByEmplId")
	ct, span := tr.Start(ctx, "repo-GetUsersByEmplId")
	defer span.End()

	rows, err := d.Db.QueryContext(ct, `SELECT users.id,
       users.display_name,
       array_to_string(array_agg(pr.name_pr), ', ')
from users
         right join user_projects up on users.id = up.user_id
         right join projects pr on up.project_id = pr.id
where empl_id = $1
group by users.id
limit 100;`, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*UserWithProjects

	for rows.Next() {
		user := &UserWithProjects{}

		err = rows.Scan(&user.Id, &user.DisplayName, &user.Projects)
		if err != nil {
			return nil, err
		}

		result = append(result, user)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Db) GetEmplList(ctx context.Context) ([]*Employee, error) {
	tr := otel.Tracer("repo-GetEmplList")
	ct, span := tr.Start(ctx, "repo-GetEmplList")
	defer span.End()

	rows, err := d.Db.QueryContext(ct, `SELECT * FROM employees limit 50`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*Employee
	for rows.Next() {
		empl := &Employee{}
		err = rows.Scan(&empl.Id, &empl.Name)
		if err != nil {
			return nil, err
		}
		result = append(result, empl)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Db) GetUser(ctx context.Context, userId int) (*User, *Employee, error) {
	tr := otel.Tracer("repo-GetUser")
	ct, span := tr.Start(ctx, "repo-GetUser")
	defer span.End()

	user := &User{}
	row := d.Db.QueryRowContext(ct, `SELECT id, display_name, empl_id, email, phone, birthday, skills from users 
                                                                 where id=$1`, userId)
	err := row.Scan(&user.Id, &user.DisplayName, &user.EmployeeId, &user.Email, &user.Phone, &user.Birthday, &user.Skills)
	if err != nil {
		return nil, nil, err
	}

	empl := &Employee{}
	row = d.Db.QueryRowContext(ct, `SELECT * from employees where id=$1`, user.EmployeeId)
	err = row.Scan(&empl.Id, &empl.Name)
	if err != nil {
		return nil, nil, err
	}

	return user, empl, nil
}

func (d *Db) AddSkillsToUserProfile(ctx context.Context, username string, email string, skills string) error {
	tr := otel.Tracer("repo-AddSkillsToUserProfile")
	ct, span := tr.Start(ctx, "repo-AddSkillsToUserProfile")
	defer span.End()

	tx, err := d.Db.BeginTx(ct, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		errOfRollback := tx.Rollback()
		if errOfRollback != nil {
			return errOfRollback
		}
		return err
	}

	row := tx.QueryRow(`SELECT skills from users where username=$1 and email=$2`, username, email)

	var oldSkills string
	err = row.Scan(&oldSkills)
	var newSkills string
	if err != nil {
		newSkills = skills
	} else {
		newSkills = fmt.Sprintf(oldSkills + "," + skills)
	}

	_, err = tx.Exec(`UPDATE users u set skills =$1 where u.username=$2 and u.email=$3`, newSkills, username, email)
	if err != nil {
		errOfRollback := tx.Rollback()
		if errOfRollback != nil {
			return errOfRollback
		}
		return errors.New("Error when adding skills")
	}

	err = tx.Commit()
	if err != nil {
		return errors.New("Error when commiting adding skills transactions")
	}

	return nil
}

func (d *Db) PutProfilesToDB(ctx context.Context, users []UserProfileFromExcel) (int, error) {
	tr := otel.Tracer("repo-PutProfilesToDB")
	ct, span := tr.Start(ctx, "repo-PutProfilesToDB")
	defer span.End()

	countOfSuccesfulTransactions := 0

	for _, user := range users {
		tx, err := d.Db.BeginTx(ct, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			errOfRollback := tx.Rollback()
			if errOfRollback != nil {
				return 0, errOfRollback
			}
			continue
		}

		emplResult := tx.QueryRow(`INSERT INTO employees (name_empl) values ($1) RETURNING id`, user.Employee)

		var emplId int
		err = emplResult.Scan(&emplId)
		if err != nil {
			errOfRollback := tx.Rollback()
			if errOfRollback != nil {
				return 0, errOfRollback
			}
			continue
		}

		userResult := tx.QueryRow(`INSERT INTO users (display_name, empl_id, email, phone, skills) values 
            ($1, $2, $3, $4, $5) RETURNING id`, user.DisplayName, emplId, user.Email, user.Phone, user.Skills)

		var userId int
		err = userResult.Scan(&userId)
		if err != nil {
			errOfRollback := tx.Rollback()
			if errOfRollback != nil {
				return 0, errOfRollback
			}
			continue
		}

		for _, device := range user.Devices {
			if device.Name == "" {
				continue
			}

			_, err := tx.Exec(`INSERT INTO devices (name_dev, type_dev, user_id) values ($1, $2, $3)`,
				device.Name, device.Type, userId)

			if err != nil {
				errOfRollback := tx.Rollback()
				if errOfRollback != nil {
					return 0, errOfRollback
				}
			}
		}

		if user.MobileDevices != nil {
			for _, device := range user.MobileDevices {
				if strings.HasPrefix(strings.ToLower(device), "iphone") {
					_, err := tx.Exec(`INSERT INTO mobile_devices (name_mob_dev, os) VALUES ($1, $2)`, device, "ios")
					if err != nil {
						errOfRollback := tx.Rollback()
						if errOfRollback != nil {
							return 0, errOfRollback
						}
					}
					continue
				}
				_, err := tx.Exec(`INSERT INTO mobile_devices (name_mob_dev, os) VALUES ($1, $2)`, device, "android")
				if err != nil {
					errOfRollback := tx.Rollback()
					if errOfRollback != nil {
						return 0, errOfRollback
					}
				}
			}
		}

		err = tx.Commit()
		if err != nil {
			errOfRollback := tx.Rollback()
			if errOfRollback != nil {
				return 0, errOfRollback
			}
			continue
		}
		countOfSuccesfulTransactions++

	}

	return countOfSuccesfulTransactions, nil
}

func (d *Db) GetUsersProfiles(ctx context.Context) ([]*UserProfile, error) {
	tr := otel.Tracer("repo-GetUsersProfiles")
	ct, span := tr.Start(ctx, "repo-GetUsersProfiles")
	defer span.End()

	rows, err := d.Db.QueryContext(ct,
		`SELECT usr.display_name,usr.phone,usr.email,usr.skills,e.name_empl,array_to_string(array_agg(d.name_dev), ', '),
       		  coalesce((SELECT array_to_string(array_agg(md.name_mob_dev), ', ')
        				from users u
            				right join renting_devices rd on u.id = rd.user_id
            				right join mobile_devices md on md.id = rd.mobile_device_id
        				where u.username=usr.username
        				group by u.username),' ')
			from users usr
    			left outer join employees e on e.id = usr.empl_id
    			left outer join devices d on usr.id = d.user_id
			group by usr.username, usr.display_name, usr.phone, usr.email, usr.skills, e.name_empl;`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*UserProfile
	for rows.Next() {
		user := &UserProfile{}
		err = rows.Scan(&user.DisplayName, &user.Phone, &user.Email, &user.Skills, &user.Employee,
			&user.Devices, &user.MobileDevices)
		if err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}
