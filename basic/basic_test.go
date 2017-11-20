package basic

import (
	"encoding/base64"
	"errors"
	"reflect"
	"testing"

	"github.com/ViBiOh/auth/auth"
	"golang.org/x/crypto/bcrypt"
)

func Test_LoadUsers(t *testing.T) {
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
		err := LoadUsers(testCase.input)
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
		want      int
	}{
		{
			`should load users from flag args`,
			1,
		},
	}

	for _, testCase := range cases {
		authUsersInput := `1:admin:admin`
		authUsers = &authUsersInput

		Auth{}.Init()
		if result := len(users); result != testCase.want {
			t.Errorf("%s\nInit() = %+v, want %+v", testCase.intention, result, testCase.want)
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
		if result := (Auth{}).GetName(); result != testCase.want {
			t.Errorf("%s\nGetName() = %+v, want %+v", testCase.intention, result, testCase.want)
		}
	}
}

func Test_GetUser(t *testing.T) {
	users = make(map[string]*basicUser)

	password, _ := bcrypt.GenerateFromPassword([]byte(`password`), 12)
	user := auth.NewUser(0, `admin`, ``)
	admin := basicUser{user, password}
	users[`admin`] = &admin

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
		result, err := Auth{}.GetUser(testCase.auth)

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

func Test_GetAccessToken(t *testing.T) {
	var cases = []struct {
		intention string
		want      error
	}{
		{
			`should return no implementation`,
			ErrNoToken,
		},
	}

	for _, testCase := range cases {
		if _, result := (Auth{}).GetAccessToken(``, ``); result != testCase.want {
			t.Errorf("%s\nGetAccessToken() = %+v, want %+v", testCase.intention, result, testCase.want)
		}
	}
}
