package cookie

import (
	"fmt"
	"net/http"
)

// GetCookieValue retrieve cookie value
func GetCookieValue(r *http.Request, name string) (string, error) {
	rawCookie, err := r.Cookie(name)
	if err != nil {
		if err != http.ErrNoCookie {
			return ``, fmt.Errorf(`Error while getting %s cookie: %v`, name, err)
		}
		return ``, nil
	}

	return rawCookie.Value, nil
}
