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

type UserPrincipal struct {
	Id       int
	Username string
	Email    string
}

// Skill represents the model for an user
type Skill struct {
	Skills string
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
