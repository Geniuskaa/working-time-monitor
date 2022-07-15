package postgres

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	Id          int              `json:"id"`
	Username    string           `json:"username"`
	DisplayName string           `json:"display_name"`
	EmployeeId  int              `json:"employee"`
	Email       string           `json:"email"`
	Phone       string           `json:"phone"`
	Birthday    pgtype.Date      `json:"birthday" swaggerignore:"true"`
	Skills      string           `json:"skills"`
	CreatedAt   pgtype.Timestamp `json:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at"`
}

type Employee struct {
	Id   int    `json:"id"`
	Name string `json:"name_empl"`
}

type UserWithProjects struct {
	Id          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Projects    string `json:"projects"`
}

type UserPrincipal struct {
	Id       int
	Username string
	Email    string
}

// Skill model info
// @Description skills which you want add to profile
type Skill struct {
	Skills string `example:"Some skills"`
}

type UserProfileFromExcel struct {
	DisplayName   string   `json:"display_name"`
	Employee      string   `json:"employee"`
	Phone         string   `json:"phone"`
	Email         string   `json:"email"`
	Devices       []Device `json:"devices"`
	MobileDevices []string `json:"mobile_devices"`
	Skills        string   `json:"skills"`
}

type Device struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// UserProfile model info
// @Description User profile information
type UserProfile struct {
	DisplayName   string `json:"display_name" example:"Зиннатуллин Эмиль Рамилевич"`
	Employee      string `json:"employee" example:"Go-developer"`
	Phone         string `json:"phone" example:"+79472738427"`
	Email         string `json:"email" example:"test@mail.ru"`
	Devices       string `json:"devices" example:""`
	MobileDevices string `json:"mobile_devices" example:"iphone 11"`
	Skills        string `json:"skills" example:"A lot of skills"`
}
