package basic

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	"github.com/ViBiOh/auth/auth"
	"golang.org/x/crypto/bcrypt"
)

func TestLoadUsers(t *testing.T) {
	var cases = []struct {
		input   string
		want    int
		wantErr error
	}{
		{
			``,
			0,
			nil,
		},
		{
			`invalid_username`,
			0,
			fmt.Errorf(`Invalid format of user for invalid_username`),
		},
		{
			`admin:admin,guest:guest`,
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
			t.Errorf(`LoadUsers(%+v) = (%+v, %+v), want (%+v, %+v)`, testCase.input, result, err, testCase.want, testCase.wantErr)
		}
	}
}

func Test_GetUser(t *testing.T) {
	users = make(map[string]*User)

	password, _ := bcrypt.GenerateFromPassword([]byte(`password`), 12)
	admin := User{`admin`, password}
	users[`admin`] = &admin

	var cases = []struct {
		auth    string
		want    *auth.User
		wantErr error
	}{
		{
			`admin`,
			nil,
			fmt.Errorf(`Error while decoding basic authentication: illegal base64 data at input byte 4`),
		},
		{
			base64.StdEncoding.EncodeToString([]byte(`AdMiN`)),
			nil,
			fmt.Errorf(`Error while reading basic authentication`),
		},
		{
			base64.StdEncoding.EncodeToString([]byte(`guest:password`)),
			nil,
			fmt.Errorf(`Invalid credentials for guest`),
		},
		{
			base64.StdEncoding.EncodeToString([]byte(`AdMiN:password`)),
			&auth.User{Username: `admin`},
			nil,
		},
	}

	var failed bool

	for _, testCase := range cases {
		result, err := GetUser(testCase.auth)

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
			t.Errorf(`GetUser(%+v) = (%+v, %+v) want (%+v, %+v)`, testCase.auth, result, err, testCase.want, testCase.wantErr)
		}
	}
}
