package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHasProfile(t *testing.T) {
	var cases = []struct {
		instance User
		profile  string
		want     bool
	}{
		{
			User{},
			`admin`,
			false,
		},
		{
			User{profiles: `admin`},
			`admin`,
			true,
		},
		{
			User{profiles: `admin,multi`},
			`multi`,
			true,
		},
		{
			User{profiles: `multi`},
			`admin`,
			false,
		},
	}

	for _, testCase := range cases {
		if result := testCase.instance.HasProfile(testCase.profile); result != testCase.want {
			t.Errorf(`%v.HasProfile(%v) = %v, want %v`, testCase.profile, testCase.instance, result, testCase.want)
		}
	}
}

func TestLoadUsersProfiles(t *testing.T) {
	var cases = []struct {
		usersAndProfiles string
		want             int
	}{
		{
			``,
			0,
		},
		{
			`admin:admin`,
			1,
		},
		{
			`admin:admin,multi|guest:|visitor:visitor`,
			3,
		},
	}

	for _, testCase := range cases {
		if result := len(LoadUsersProfiles(testCase.usersAndProfiles)); result != testCase.want {
			t.Errorf(`LoadUsersProfiles(%v) = %v, want %v`, testCase.usersAndProfiles, result, testCase.want)
		}
	}
}

func TestIsAuthenticated(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Authorization`) == `unauthorized` {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.Write([]byte(r.Header.Get(`Authorization`)))
		}
	}))
	defer testServer.Close()

	admin := NewUser(`admin`, `admin`)

	var cases = []struct {
		authorization string
		want          *User
		wantErr       error
	}{
		{
			`unauthorized`,
			nil,
			fmt.Errorf(`Error while getting username: Error status 401: `),
		},
		{
			`guest`,
			nil,
			fmt.Errorf(`[guest] Not allowed to use app`),
		},
		{
			`admin`,
			admin,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		req := httptest.NewRequest(http.MethodGet, testServer.URL, nil)
		req.Header.Set(authorizationHeader, testCase.authorization)
		result, err := IsAuthenticated(testServer.URL, map[string]*User{`admin`: admin}, req)

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if result != testCase.want {
			failed = true
		}

		if failed {
			t.Errorf(`IsAuthenticated(%v) = (%v, %v), want (%v, %v)`, req, result, err, testCase.want, testCase.wantErr)
		}
	}
}
func TestIsAuthenticatedByAuth(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Authorization`) == `unauthorized` {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.Write([]byte(r.Header.Get(`Authorization`)))
		}
	}))
	defer testServer.Close()

	admin := NewUser(`admin`, `admin`)

	var cases = []struct {
		authorization string
		want          *User
		wantErr       error
	}{
		{
			`unauthorized`,
			nil,
			fmt.Errorf(`Error while getting username: Error status 401: `),
		},
		{
			`guest`,
			nil,
			fmt.Errorf(`[guest] Not allowed to use app`),
		},
		{
			`admin`,
			admin,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := IsAuthenticatedByAuth(testServer.URL, map[string]*User{`admin`: admin}, testCase.authorization, `127.0.0.1`)

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if result != testCase.want {
			failed = true
		}

		if failed {
			t.Errorf(`IsAuthenticatedByAuth(%v) = (%v, %v), want (%v, %v)`, testCase.authorization, result, err, testCase.want, testCase.wantErr)
		}
	}
}
