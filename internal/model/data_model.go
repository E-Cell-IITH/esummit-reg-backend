package model

type RegistrationData struct {
	SName    string `json:"sname"`
	FName    string `json:"fname"`
	POCName  string `json:"pocname"`
	Contact  string `json:"contact"`
	Startup  string `json:"startup"`
	Service  string `json:"service"`
	Email    string `json:"email"`
	SEmail   string `json:"semail"`
	IFocus   string `json:"ifocus"`
	AYears   string `json:"ayears"`
	Location string `json:"location"`
	City     string `json:"city"`
	About    string `json:"about"`
}

type RegistrationRequest struct {
	Data  RegistrationData `json:"data"`
	Token string           `json:"token"`
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Data  string `json:"data"`
}