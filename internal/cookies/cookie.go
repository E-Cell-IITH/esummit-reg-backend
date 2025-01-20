package cookies

// import (
// 	"net/http"
// 	"time"
// )

// var age = 60 * 60 * 24 * 5

// func SetCookie(w http.ResponseWriter, name, value string, maxAge int) {
// 	if maxAge == 0 {
// 		maxAge = age
// 	}
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     name,
// 		Value:    value,
// 		Path:     "/",
// 		Domain:   ".ecelliith.org.in",
// 		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
// 		MaxAge:   maxAge,
// 		HttpOnly: true,
// 		Secure:   true,
// 		SameSite: http.SameSiteNoneMode,
// 	})
// }
