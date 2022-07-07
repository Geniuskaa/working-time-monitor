package device

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
	Id           int              `json:"id"`
	MobileDevice MobileDevice     `json:"mobileDevice"`
	UserId       int              `json:"userId"`
	CreatedAt    pgtype.Timestamp `json:"createdAt"`
	UpdatedAt    pgtype.Timestamp `json:"updatedAt"`
}

type MobileDevice struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Os   string `json:"os"`
}
