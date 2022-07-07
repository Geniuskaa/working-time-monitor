package device

type RentingDeviceResponse struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Free bool   `json:"free"`
}
