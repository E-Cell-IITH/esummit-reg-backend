package model

type RegistrationData struct {
	Id       *int   `json:"id"`
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
	ID            int    `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	ContactNumber string `json:"contact_number"`
	Data          string `json:"data"`
}

type PurchasedTicketWithUser struct {
	ID              int64   `json:"id"`
	UserID          int64   `json:"user_id"`
	TicketTitle     string  `json:"ticket_title"`
	Price           float64 `json:"price"`
	IsAccommodation bool    `json:"is_accommodation"`
	Coupon          string  `json:"coupon"`
	User            User    `json:"user"`
}

