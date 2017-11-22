package cookie

import (
	"net/http"
)

// GetCookieValue retrieve cookie value
func GetCookieValue(r *http.Request, name string) string {
	rawCookie, err := r.Cookie(name)
	if err == http.ErrNoCookie {
		return ``
	}

	return rawCookie.Value
}
