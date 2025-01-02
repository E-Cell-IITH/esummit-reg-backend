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
	IDToken  string `json:"id_token"`
}

type RegistrationRequest struct {
	Data  RegistrationData `json:data`
	Token string           `json:token`
}
