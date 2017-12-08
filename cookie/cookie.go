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

// SetCookieAndRedirect defined auth cookie and redirect
func SetCookieAndRedirect(w http.ResponseWriter, r *http.Request, redirect string, cookieDomain string, cookieContent string) {
	http.SetCookie(w, &http.Cookie{
		Name:     `auth`,
		Path:     `/`,
		Value:    cookieContent,
		Domain:   cookieDomain,
		Secure:   true,
		HttpOnly: true,
	})

	http.Redirect(w, r, redirect, http.StatusFound)
}

// ClearCookieAndRedirect drop existing cookie for auth
func ClearCookieAndRedirect(w http.ResponseWriter, r *http.Request, redirect string, cookieDomain string) {
	http.SetCookie(w, &http.Cookie{
		Name:     `auth`,
		Path:     `/`,
		Value:    `none`,
		Domain:   cookieDomain,
		Secure:   true,
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, redirect, http.StatusFound)
}
