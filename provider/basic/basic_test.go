package basic

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ViBiOh/auth/auth"
	"golang.org/x/crypto/bcrypt"
)

func Test_Flags(t *testing.T) {
	var cases = []struct {
		intention string
		want      int
	}{
		{
			`should return map with one entries`,
			1,
		},
	}

	for _, testCase := range cases {
		if result := Flags(``); len(result) != testCase.want {
			t.Errorf("%s\nFlags() = %+v, want %+v", testCase.intention, result, testCase.want)
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
			errors.New(`Invalid format of user for invalid_username`),
		},
		{
			`should handle invalid uint format`,
			`abc:invalid_username:abc`,
			0,
			errors.New(`Invalid id format for user abc:invalid_username:abc`),
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

func Test_Init(t *testing.T) {
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
			fmt.Errorf(`Error while loading users: Invalid format of user for invalid format`),
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
		authClient := Auth{}
		err := authClient.Init(map[string]interface{}{`users`: &testCase.users})

		failed = false

		if err == nil && testCase.wantErr != nil {
			failed = true
		} else if err != nil && testCase.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != testCase.wantErr.Error() {
			failed = true
		} else if len(authClient.users) != testCase.want {
			failed = true
		}

		if failed {
			t.Errorf("%s\nInit(%+v) = (%+v, %+v), want (%+v, %+v)", testCase.intention, testCase.users, authClient.users, err, testCase.want, testCase.wantErr)
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
	authClient.users = map[string]*basicUser{`admin`: &basicUser{auth.NewUser(0, `admin`, ``), password}}

	var cases = []struct {
		intention string
		auth      string
		want      *auth.User
		wantErr   error
	}{
		{
			`should handle malformed header`,
			`admin`,
			nil,
			errors.New(`Error while decoding basic authentication: illegal base64 data at input byte 4`),
		},
		{
			`should handle malformed content`,
			base64.StdEncoding.EncodeToString([]byte(`AdMiN`)),
			nil,
			errors.New(`Error while reading basic authentication`),
		},
		{
			`should handle not found user`,
			base64.StdEncoding.EncodeToString([]byte(`guest:password`)),
			nil,
			errors.New(`Invalid credentials for guest`),
		},
		{
			`should handle invalid credentials`,
			base64.StdEncoding.EncodeToString([]byte(`AdMiN:admin`)),
			nil,
			errors.New(`Invalid credentials for admin`),
		},
		{
			`should handle valid auth`,
			base64.StdEncoding.EncodeToString([]byte(`AdMiN:password`)),
			&auth.User{ID: 0, Username: `admin`},
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := authClient.GetUser(testCase.auth)

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
		if result, _ := (&Auth{}).Redirect(); result != testCase.want {
			t.Errorf("%s\nRedirect() = (%+v), want (%+v)", testCase.intention, result, testCase.want)
		}
	}
}
