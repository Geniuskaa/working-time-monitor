package user

type EmpolyeeDTO struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type UserWithProjectsDTO struct {
	Id          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Projects    string `json:"projects"`
}
