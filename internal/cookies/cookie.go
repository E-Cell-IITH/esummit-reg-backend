package cookies

import (
	"net/http"
	"time"
)

var maxAge = 60 * 60 * 24 * 1

func SetCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   "register.ecelliith.org.in",
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}
