package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ViBiOh/httputils"
)

func authTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Authorization`) == `` {
			http.Error(w, ErrEmptyAuthorization.Error(), http.StatusUnauthorized)
		} else if r.Header.Get(`Authorization`) == `unauthorized` {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.Write([]byte(r.Header.Get(`Authorization`)))
		}
	}))
}

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

func Test_IsForbiddenErr(t *testing.T) {
	var cases = []struct {
		intention string
		err       error
		want      bool
	}{
		{
			`should identify error with pattern`,
			fmt.Errorf(`An error occurred %s`, forbiddenMessage),
			true,
		},
		{
			`should identify error without pattern`,
			errors.New(`Not allowed`),
			false,
		},
	}

	for _, testCase := range cases {
		if result := IsForbiddenErr(testCase.err); result != testCase.want {
			t.Errorf("%s\nIsForbiddenErr(%+v) = %+v, want %+v", testCase.intention, testCase.err, result, testCase.want)
		}
	}
}

func Test_IsAuthenticated(t *testing.T) {
	testServer := authTestServer()
	defer testServer.Close()

	admin := NewUser(1, `admin`, `admin`)

	var cases = []struct {
		intention     string
		authorization string
		want          *User
		wantErr       error
	}{
		{
			`should forward header`,
			`{"id":1,"username":"admin"}`,
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
			t.Errorf("%s\nIsAuthenticated(%+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.authorization, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func Test_IsAuthenticatedByAuth(t *testing.T) {
	testServer := authTestServer()
	defer testServer.Close()

	admin := NewUser(1, `admin`, `admin`)

	var cases = []struct {
		intention     string
		authorization string
		want          *User
		wantErr       error
	}{
		{
			`should handle unauthorized user`,
			`unauthorized`,
			nil,
			errors.New(`Error while getting user: Error status 401`),
		},
		{
			`should handle empty header`,
			``,
			nil,
			ErrEmptyAuthorization,
		},
		{
			`should handle forbidden user`,
			`{"id":8000,"username":"guest"}`,
			nil,
			errors.New(`[guest] Not allowed to use app`),
		},
		{
			`should handle invalid user format`,
			`{"id":1,"username":"admin"`,
			nil,
			errors.New(`Error while unmarshalling user: unexpected end of JSON input`),
		},
		{
			`should handle valid user`,
			`{"id":1,"username":"admin"}`,
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
			t.Errorf("%s\nIsAuthenticatedByAuth(%+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.authorization, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func Test_Handler(t *testing.T) {
	testServer := authTestServer()
	defer testServer.Close()

	admin := NewUser(1, `admin`, `admin`)
	next := func(w http.ResponseWriter, _ *http.Request, user *User) {
		httputils.ResponseJSON(w, http.StatusOK, user, false)
	}

	handler := Handler(testServer.URL, map[string]*User{`admin`: admin}, next)

	var cases = []struct {
		intention     string
		authorization string
		want          string
		wantStatus    int
	}{
		{
			`should handle empty authorization header`,
			``,
			`Empty authorization content
`,
			http.StatusUnauthorized,
		},
		{
			`should handle unauthorized user`,
			`unauthorized`,
			`Error while getting user: Error status 401
`,
			http.StatusUnauthorized,
		},
		{
			`should handle forbidden user`,
			`{"id":8000,"username":"guest"}`,
			`⛔️
`,
			http.StatusForbidden,
		},
		{
			`should call next handler`,
			`{"id":1,"username":"admin"}`,
			`{"id":1,"username":"admin"}`,
			http.StatusOK,
		},
	}

	for _, testCase := range cases {
		req := httptest.NewRequest(http.MethodGet, testServer.URL, nil)
		req.Header.Set(authorizationHeader, testCase.authorization)
		writer := httptest.NewRecorder()

		handler.ServeHTTP(writer, req)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%v\nHandler(%+v) = %+v, want status %+v", testCase.intention, testCase.authorization, result, testCase.wantStatus)
		}

		if result, _ := httputils.ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf("%s\nHandler(%+v) = %+v, want %+v", testCase.intention, testCase.authorization, string(result), testCase.want)
		}
	}
}
