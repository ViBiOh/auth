package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/auth"
	"github.com/ViBiOh/auth/pkg/cookie"
	"github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
)

// GetUser get user from given auth content
func (a *App) GetUser(authContent string) (*provider.User, error) {
	if authContent == `` {
		return nil, auth.ErrEmptyAuthorization
	}

	parts := strings.SplitN(authContent, ` `, 2)
	if len(parts) != 2 {
		return nil, provider.ErrMalformedAuth
	}

	for _, provider := range a.providers {
		if parts[0] == provider.GetName() {
			user, err := provider.GetUser(parts[1])
			if err != nil {
				return nil, err
			}
			return user, nil
		}
	}

	return nil, provider.ErrUnknownAuthType
}

func (a *App) userHandler(w http.ResponseWriter, r *http.Request) {
	user, err := a.GetUser(auth.ReadAuthContent(r))
	if err != nil {
		if err == provider.ErrMalformedAuth || err == provider.ErrUnknownAuthType {
			httperror.BadRequest(w, err)
		} else {
			httperror.Unauthorized(w, err)
		}

		return
	}

	if err := httpjson.ResponseJSON(w, http.StatusOK, user, httpjson.IsPretty(r.URL.RawQuery)); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a *App) redirectHandler(w http.ResponseWriter, r *http.Request) {
	for _, provider := range a.providers {
		if strings.HasSuffix(r.URL.Path, strings.ToLower(provider.GetName())) {
			if redirect, err := provider.Redirect(); err != nil {
				httperror.InternalServerError(w, err)
			} else {
				http.Redirect(w, r, redirect, http.StatusFound)
			}

			return
		}
	}

	httperror.BadRequest(w, provider.ErrUnknownAuthType)
}

func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	for _, provider := range a.providers {
		if strings.HasSuffix(r.URL.Path, strings.ToLower(provider.GetName())) {
			if token, err := provider.Login(r); err != nil {
				w.Header().Add(`WWW-Authenticate`, fmt.Sprintf(`%s charset="UTF-8"`, provider.GetName()))
				httperror.Unauthorized(w, err)
			} else if a.redirect != `` {
				cookie.SetCookieAndRedirect(w, r, a.redirect, a.cookieDomain, fmt.Sprintf(`%s %s`, provider.GetName(), token))
			} else if _, err := w.Write([]byte(token)); err != nil {
				httperror.InternalServerError(w, err)
			}

			return
		}
	}

	httperror.BadRequest(w, provider.ErrUnknownAuthType)
}

func (a *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie.ClearCookieAndRedirect(w, r, a.redirect, a.cookieDomain)
}
