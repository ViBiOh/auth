package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/auth/pkg/auth"
	"github.com/ViBiOh/auth/pkg/cookie"
	"github.com/ViBiOh/auth/pkg/model"
	"github.com/ViBiOh/auth/pkg/provider"
	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
)

// GetUser get user from given auth content
func (a App) GetUser(ctx context.Context, authContent string) (*model.User, error) {
	if authContent == `` {
		return nil, provider.ErrEmptyAuthorization
	}

	parts := strings.SplitN(authContent, ` `, 2)
	if len(parts) != 2 {
		return nil, provider.ErrMalformedAuth
	}

	for _, p := range a.providers {
		if parts[0] == p.GetName() {
			user, err := p.GetUser(ctx, parts[1])
			if err != nil {
				return nil, err
			}

			return user, nil
		}
	}

	return nil, provider.ErrUnknownAuthType
}

func (a App) userHandler(w http.ResponseWriter, r *http.Request) {
	user, err := a.GetUser(r.Context(), auth.ReadAuthContent(r))
	if err != nil {
		if err == provider.ErrMalformedAuth || err == provider.ErrUnknownAuthType {
			httperror.BadRequest(w, err)
		} else {
			httperror.Unauthorized(w, err)
		}

		return
	}

	if err := httpjson.ResponseJSON(w, http.StatusOK, user, httpjson.IsPretty(r)); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a App) redirectHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.TrimPrefix(r.URL.Path, redirectPrefix)

	if url == `` || url == `/` {
		a.RedirectToFirstProvider(w, r)
		return
	}

	for _, p := range a.providers {
		if strings.HasSuffix(url, strings.ToLower(p.GetName())) {
			p.Redirect(w, r)
			return
		}
	}

	httperror.BadRequest(w, provider.ErrUnknownAuthType)
}

// RedirectToFirstProvider redirects user to first provider in list
func (a App) RedirectToFirstProvider(w http.ResponseWriter, r *http.Request) bool {
	if len(a.providers) > 0 {
		a.providers[0].Redirect(w, r)
		return true
	}

	return false
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

	httperror.BadRequest(w, provider.ErrUnknownAuthType)
}

func (a App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie.ClearCookieAndRedirect(w, r, a.redirect, a.cookieDomain)
}
