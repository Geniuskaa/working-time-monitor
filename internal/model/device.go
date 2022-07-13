package model

import "github.com/jackc/pgx/v5/pgtype"

type Device struct {
	Id        int              `json:"id"`
	Name      string           `json:"name"`
	Type      string           `json:"type"`
	UserId    int              `json:"userId"`
	CreatedAt pgtype.Timestamp `json:"createdAt"`
	UpdatedAt pgtype.Timestamp `json:"updatedAt"`
}

type RentingDevice struct {
	Id           int               `json:"id"`
	MobileDevice MobileDevice      `json:"mobileDevice"`
	User         RentingDeviceUser `json:"user"`
	CreatedAt    pgtype.Timestamp  `json:"createdAt"`
	UpdatedAt    pgtype.Timestamp  `json:"updatedAt"`
}

type RentingDeviceUser struct {
	Id          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type MobileDevice struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Os   string `json:"os"`
}
