package device

// RentingDeviceResponse model info
// @Description Information about the current status of device
type RentingDeviceResponse struct {
	Id          int    `json:"id" example:"1"`
	Name        string `json:"name" example:"Iphone 12 Pro"`
	DisplayName string `json:"display_name" example:"Aydar Ibragimov"`
}
