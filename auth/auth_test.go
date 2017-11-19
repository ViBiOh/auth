package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_HasProfile(t *testing.T) {
	var cases = []struct {
		intention string
		instance  User
		profile   string
		want      bool
	}{
		{
			`should handle nil profiles`,
			User{},
			`admin`,
			false,
		},
		{
			`should find simple match`,
			User{profiles: `admin`},
			`admin`,
			true,
		},
		{
			`should find match when multiples values`,
			User{profiles: `admin|multi`},
			`multi`,
			true,
		},
		{
			`should find no match`,
			User{profiles: `multi`},
			`admin`,
			false,
		},
	}

	for _, testCase := range cases {
		if result := testCase.instance.HasProfile(testCase.profile); result != testCase.want {
			t.Errorf("%s\n%+v.HasProfile(%+v) = %+v, want %+v", testCase.intention, testCase.instance, testCase.profile, result, testCase.want)
		}
	}
}

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		prefix    string
		want      int
	}{
		{
			`should return map with two entries`,
			``,
			2,
		},
	}

	for _, testCase := range cases {
		if result := Flags(testCase.prefix); len(result) != testCase.want {
			t.Errorf("%s\nFlags(%+v) = %+v, want %+v", testCase.intention, testCase.prefix, result, testCase.want)
		}
	}
}

func Test_NewUser(t *testing.T) {
	var cases = []struct {
		intention string
		id        uint
		username  string
		profiles  string
		want      *User
	}{
		{
			`should work with given params`,
			1,
			`vibioh`,
			`admin|multi`,
			&User{1, `vibioh`, `admin|multi`},
		},
	}

	for _, testCase := range cases {
		if result := NewUser(testCase.id, testCase.username, testCase.profiles); !reflect.DeepEqual(result, testCase.want) {
			t.Errorf("%s\nNewUser(%+v, %+v, %+v) = %+v, want %+v", testCase.intention, testCase.id, testCase.username, testCase.profiles, result, testCase.want)
		}
	}
}

func Test_LoadUsersProfiles(t *testing.T) {
	var cases = []struct {
		intention        string
		usersAndProfiles string
		want             int
	}{
		{
			`should handle empty string`,
			``,
			0,
		},
		{
			`should handle one user`,
			`admin:admin`,
			1,
		},
		{
			`should handle multiples users`,
			`admin:admin|multi,guest:,visitor:visitor`,
			3,
		},
	}

	for _, testCase := range cases {
		if result := len(LoadUsersProfiles(testCase.usersAndProfiles)); result != testCase.want {
			t.Errorf("%s\nLoadUsersProfiles(%+v) = %+v, want %+v", testCase.intention, testCase.usersAndProfiles, result, testCase.want)
		}
	}
}

func Test_IsAuthenticated(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Authorization`) == `unauthorized` {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.Write([]byte(r.Header.Get(`Authorization`)))
		}
	}))
	defer testServer.Close()

	admin := NewUser(0, `admin`, `admin`)

	var cases = []struct {
		authorization string
		want          *User
		wantErr       error
	}{
		{
			`unauthorized`,
			nil,
			fmt.Errorf(`Error while getting user: Error status 401`),
		},
		{
			`{"id":8000,"username":"guest"}`,
			nil,
			fmt.Errorf(`[guest] Not allowed to use app`),
		},
		{
			`{"id":8000,"username":"admin"}`,
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
			t.Errorf(`IsAuthenticated(%+v) = (%+v, %+v), want (%+v, %+v)`, req, result, err, testCase.want, testCase.wantErr)
		}
	}
}
func Test_IsAuthenticatedByAuth(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Authorization`) == `unauthorized` {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.Write([]byte(r.Header.Get(`Authorization`)))
		}
	}))
	defer testServer.Close()

	admin := NewUser(0, `admin`, `admin`)

	var cases = []struct {
		authorization string
		want          *User
		wantErr       error
	}{
		{
			`unauthorized`,
			nil,
			fmt.Errorf(`Error while getting user: Error status 401`),
		},
		{
			`{"id":8000,"username":"guest"}`,
			nil,
			fmt.Errorf(`[guest] Not allowed to use app`),
		},
		{
			`{"id":100,"username":"admin"}`,
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
			t.Errorf(`IsAuthenticatedByAuth(%+v) = (%+v, %+v), want (%+v, %+v)`, testCase.authorization, result, err, testCase.want, testCase.wantErr)
		}
	}
}
