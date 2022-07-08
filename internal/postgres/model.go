package postgres

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Skill struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	Id          int              `json:"id"`
	Username    string           `json:"username"`
	DisplayName string           `json:"display_name"`
	EmployeeId  int              `json:"employee"`
	Email       string           `json:"email"`
	Phone       string           `json:"phone"`
	Birthday    pgtype.Date      `json:"birthday"`
	Skills      string           `json:"skills"`
	CreatedAt   pgtype.Timestamp `json:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at"`
}

type Employee struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type UserWithProjects struct {
	Id          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Projects    string `json:"projects"`
}
