package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
		wantType  string
	}{
		{
			`should add string url param to flags`,
			`url`,
			`*string`,
		},
		{
			`should add string users param to flags`,
			`users`,
			`*string`,
		},
	}

	for _, testCase := range cases {
		result := Flags(testCase.intention)[testCase.want]

		if result == nil {
			t.Errorf("%s\nFlags() = %+v, want `%s`", testCase.intention, result, testCase.want)
		}

		if fmt.Sprintf(`%T`, result) != testCase.wantType {
			t.Errorf("%s\nFlags() = `%T`, want `%s`", testCase.intention, result, testCase.wantType)
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

func Test_loadUsersProfiles(t *testing.T) {
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
		if result := len(loadUsersProfiles(testCase.usersAndProfiles)); result != testCase.want {
			t.Errorf("%s\nloadUsersProfiles(%+v) = %+v, want %+v", testCase.intention, testCase.usersAndProfiles, result, testCase.want)
		}
	}
}
