package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type UserRepo interface {
	GetUsersByEmplId(ctx context.Context, id int) ([]*UserWithProjects, error)
	GetEmplList(ctx context.Context) ([]*Employee, error)
	GetUserByUserId(ctx context.Context, id int) (User, error)
	GetUserPrincipalByUsername(ctx context.Context, username string) (*UserPrincipal, error)
	AddSkillsToUserProfile(ctx context.Context, username string, email string, skills string) error
	PutProfilesToDB(ctx context.Context, users []UserProfileFromExcel) (int, error)
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

	rows, err := d.Db.QueryContext(ctx, `SELECT users.id,
       users.display_name,
       array_to_string(array_agg(pr.name), ', ')
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
	rows, err := d.Db.QueryContext(ctx, `SELECT * FROM employees limit 50`)
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

func (d *Db) GetUserByUserId(ctx context.Context, id int) (*User, *Employee, error) {
	user := &User{}
	row := d.Db.QueryRowContext(ctx, `SELECT id, display_name, empl_id, email, phone, birthday, skills from users where id=$1`, id)
	err := row.Scan(&user.Id, &user.DisplayName, &user.EmployeeId, &user.Email, &user.Phone, &user.Birthday, &user.Skills)
	if err != nil {
		return nil, nil, err
	}

	empl := &Employee{}
	err = d.Db.GetContext(ctx, empl, `SELECT * from employees where id=$1`, user.EmployeeId)
	if err != nil {
		return nil, nil, err
	}

	return user, empl, nil
}

func (d *Db) AddSkillsToUserProfile(ctx context.Context, username string, email string, skills string) error {
	tx, err := d.Db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
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
	if err != nil {
		errOfRollback := tx.Rollback()
		if errOfRollback != nil {
			return errOfRollback
		}
		return err
	}

	newSkills := fmt.Sprintf(oldSkills + "," + skills)

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
	countOfSuccesfulTransactions := 0

	for _, user := range users {
		tx, err := d.Db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			errOfRollback := tx.Rollback()
			if errOfRollback != nil {
				return 0, errOfRollback
			}
			continue
		}

		emplResult := tx.QueryRow(`INSERT INTO employees (name) values ($1) RETURNING id`, user.Employee)

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

			_, err := tx.Exec(`INSERT INTO devices (name, type, user_id) values ($1, $2, $3)`,
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
					_, err := tx.Exec(`INSERT INTO mobile_devices (name, os) VALUES ($1, $2)`, device, "ios")
					if err != nil {
						errOfRollback := tx.Rollback()
						if errOfRollback != nil {
							return 0, errOfRollback
						}
					}
					continue
				}
				_, err := tx.Exec(`INSERT INTO mobile_devices (name, os) VALUES ($1, $2)`, device, "android")
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
