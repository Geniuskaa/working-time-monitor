package postgres

import (
	"context"
)

type UserRepo interface {
	GetUsersByEmplId(ctx context.Context, id int) ([]*UserWithProjects, error)
	GetEmplList(ctx context.Context) ([]*Employee, error)
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
