package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/auth"
	"github.com/ViBiOh/auth/pkg/cookie"
	"github.com/ViBiOh/auth/pkg/ident"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
)

func (a App) userHandler(w http.ResponseWriter, r *http.Request) {
	authContent := auth.ReadAuthContent(r)
	if authContent == `` {
		httperror.Unauthorized(w, ident.ErrEmptyAuth)
		return
	}

	parts := strings.SplitN(authContent, ` `, 2)
	if len(parts) != 2 {
		httperror.BadRequest(w, ident.ErrMalformedAuth)
		return
	}

	for _, provider := range a.providers {
		if parts[0] == provider.GetName() {
			user, err := provider.GetUser(r.Context(), parts[1])
			if err != nil {
				httperror.Unauthorized(w, err)
			}

			if err := httpjson.ResponseJSON(w, http.StatusOK, user, httpjson.IsPretty(r)); err != nil {
				httperror.InternalServerError(w, err)
			}

			return
		}
	}

	httperror.BadRequest(w, ident.ErrUnknownIdentType)
}

func (a App) redirectHandler(w http.ResponseWriter, r *http.Request) {
	for _, p := range a.providers {
		if strings.HasSuffix(r.URL.Path, strings.ToLower(p.GetName())) {
			p.Redirect(w, r)
			return
		}
	}

	httperror.BadRequest(w, ident.ErrUnknownIdentType)
}

func (a App) loginHandler(w http.ResponseWriter, r *http.Request) {
	for _, p := range a.providers {
		providerName := p.GetName()

		if strings.HasSuffix(r.URL.Path, strings.ToLower(providerName)) {
			token, err := p.Login(r)
			if err != nil {
				p.OnLoginError(w, r, err)
				return
			}

			if a.redirect != `` {
				cookie.SetCookieAndRedirect(w, r, a.redirect, a.cookieDomain, fmt.Sprintf(`%s %s`, providerName, token))
				return
			}

			if _, err := w.Write([]byte(token)); err != nil {
				httperror.InternalServerError(w, err)
				return
			}

			return
		}
	}

	httperror.BadRequest(w, ident.ErrUnknownIdentType)
}

func (a App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie.ClearCookieAndRedirect(w, r, a.redirect, a.cookieDomain)
}
