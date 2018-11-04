package basic

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ViBiOh/auth/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
		wantType  string
	}{
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

func Test_loadUsers(t *testing.T) {
	var cases = []struct {
		intention string
		input     string
		want      int
		wantErr   error
	}{
		{
			`should handle empty string`,
			``,
			0,
			nil,
		},
		{
			`should handle invalid format`,
			`invalid_username`,
			0,
			errors.New(`invalid format of user for invalid_username`),
		},
		{
			`should handle invalid uint format`,
			`abc:invalid_username:abc`,
			0,
			errors.New(`invalid id format for user abc:invalid_username:abc`),
		},
		{
			`should handle valid format`,
			`0:admin:admin,1:guest:guest`,
			2,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		users, err := loadUsers(testCase.input)
		result := len(users)

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
			t.Errorf("%s\nLoadUsers(%+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.input, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func Test_NewAuth(t *testing.T) {
	var cases = []struct {
		intention string
		users     string
		want      int
		wantErr   error
	}{
		{
			`should handle load error`,
			`invalid format`,
			0,
			errors.New(`invalid format of user for invalid format`),
		},
		{
			`should load users from given args`,
			`1:admin:admin`,
			1,
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		auth, err := NewAuth(map[string]interface{}{`users`: &testCase.users}, nil)
		var authClient *Auth
		if auth != nil {
			authClient = auth.(*Auth)
		}

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if authClient != nil && len(authClient.users) != testCase.want {
			failed = true
		}

		if failed {
			t.Errorf("%s\nNewAuth(%+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.users, authClient.users, err, testCase.want, testCase.wantErr)
		}
	}
}

func Test_GetName(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			`should return constant`,
			`Basic`,
		},
	}

	for _, testCase := range cases {
		if result := (&Auth{}).GetName(); result != testCase.want {
			t.Errorf("%s\nGetName() = %+v, want %+v", testCase.intention, result, testCase.want)
		}
	}
}

func Test_GetUser(t *testing.T) {
	password, _ := bcrypt.GenerateFromPassword([]byte(`password`), 12)
	authClient := Auth{}
	authClient.users = map[string]*basicUser{`admin`: {model.NewUser(0, `admin`, ``, ``), password}}

	var cases = []struct {
		intention string
		auth      string
		want      *model.User
		wantErr   error
	}{
		{
			`should handle malformed header`,
			`admin`,
			nil,
			errors.New(`illegal base64 data at input byte 4`),
		},
		{
			`should handle malformed content`,
			base64.StdEncoding.EncodeToString([]byte(`AdMiN`)),
			nil,
			errors.New(`invalid format for basic auth`),
		},
		{
			`should handle not found user`,
			base64.StdEncoding.EncodeToString([]byte(`guest:password`)),
			nil,
			errors.New(`invalid credentials`),
		},
		{
			`should handle invalid credentials`,
			base64.StdEncoding.EncodeToString([]byte(`AdMiN:admin`)),
			nil,
			errors.New(`invalid credentials`),
		},
		{
			`should handle valid auth`,
			base64.StdEncoding.EncodeToString([]byte(`AdMiN:password`)),
			&model.User{ID: 0, Username: `admin`},
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := authClient.GetUser(nil, testCase.auth)

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if !reflect.DeepEqual(result, testCase.want) {
			failed = true
		}

		if failed {
			t.Errorf("%s\nGetUser(%+v) = (%+v, %+v) want (%+v, %+v)", testCase.intention, testCase.auth, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func Test_Redirect(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			`should return Basic redirection`,
			`/login/basic`,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()

		(&Auth{}).Redirect(writer, httptest.NewRequest(http.MethodGet, `/`, nil))
		result := writer.Header().Get(`location`)
		if result != testCase.want {
			t.Errorf("%s\nRedirect() = (%+v), want (%+v)", testCase.intention, result, testCase.want)
		}
	}
}
