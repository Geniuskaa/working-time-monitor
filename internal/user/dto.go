package user

import "github.com/jackc/pgx/v5/pgtype"

// EmpolyeeDTO represents the model for an user
type EmpolyeeDTO struct {
	Id   int    `json:"id" example:"1"`
	Name string `json:"name" example:"Go-developer"`
}

// UserWithProjectsDTO  model info
// @Description Info about user and his projects
type UserWithProjectsDTO struct {
	Id          int    `json:"id" example:"1"`
	DisplayName string `json:"display_name" example:"Зиннатуллин Эмиль Рамилевич"`
	Projects    string `json:"projects" example:"Халвёнок, SCB-monitor"`
}

// UserDTO  model info
// @Description Main info about user and his projects
type UserDTO struct {
	Id          int         `json:"id" example:"1"`
	DisplayName string      `json:"display_name" example:"Зиннатуллин Эмиль Рамилевич"`
	Employee    string      `json:"employee" example:"Go-developer"`
	Email       string      `json:"email" example:"test@mail.ru"`
	Phone       string      `json:"phone" example:"+79648246372"`
	Birthday    pgtype.Date `json:"birthday" swaggerignore:"true"`
	Skills      string      `json:"skills" example:"A lot of skills"`
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
