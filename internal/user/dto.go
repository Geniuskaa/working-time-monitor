package user

import "github.com/jackc/pgx/v5/pgtype"

// EmpolyeeDTO represents the model for an user
type EmpolyeeDTO struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// UserWithProjectsDTO represents the model for an user
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

type UserProfileDTO struct {
	Id          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Employee    string `json:"employee"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Devices     string `json:"devices"`
	Skills      string `json:"skills"`
}
