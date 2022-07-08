package user

import "github.com/jackc/pgx/v5/pgtype"

type EmpolyeeDTO struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type UserWithProjectsDTO struct {
	Id          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Projects    string `json:"projects"`
}

type UserDTO struct {
	Id          int         `json:"id"`
	DisplayName string      `json:"display_name"`
	Employee    string      `json:"employee"`
	Email       string      `json:"email"`
	Phone       string      `json:"phone"`
	Birthday    pgtype.Date `json:"birthday"`
	Skills      string      `json:"skills"`
}
