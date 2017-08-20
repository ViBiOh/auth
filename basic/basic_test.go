package basic

import (
	"encoding/base64"
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestLoadUsers(t *testing.T) {
	var tests = []struct {
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

	for _, test := range tests {
		err := LoadUsers(test.input)
		result := len(users)

		failed = false

		if err == nil && test.wantErr != nil {
			failed = true
		} else if err != nil && test.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != test.wantErr.Error() {
			failed = true
		} else if result != test.want {
			failed = true
		}

		if failed {
			t.Errorf(`LoadUsers(%v) = (%v, %v), want (%v, %v)`, test.input, result, err, test.want, test.wantErr)
		}
	}
}

func TestGetUsername(t *testing.T) {
	users = make(map[string]*User)

	password, _ := bcrypt.GenerateFromPassword([]byte(`password`), 12)
	admin := User{`admin`, password}
	users[`admin`] = &admin

	guest, _ := bcrypt.GenerateFromPassword([]byte(`guest`), 12)
	users[`guest`] = &User{`guest`, guest}

	var tests = []struct {
		auth    string
		want    string
		wantErr error
	}{
		{
			`admin`,
			``,
			fmt.Errorf(`Error while decoding basic authentication: illegal base64 data at input byte 4`),
		},
		{
			base64.StdEncoding.EncodeToString([]byte(`AdMiN`)),
			``,
			fmt.Errorf(`Error while reading basic authentication`),
		},
		{
			base64.StdEncoding.EncodeToString([]byte(`AdMiN:password`)),
			`AdMiN`,
			nil,
		},
	}

	var failed bool

	for _, test := range tests {
		result, err := GetUsername(test.auth)

		failed = false

		if err == nil && test.wantErr != nil {
			failed = true
		} else if err != nil && test.wantErr == nil {
			failed = true
		} else if err != nil && err.Error() != test.wantErr.Error() {
			failed = true
		} else if result != test.want {
			failed = true
		}

		if failed {
			t.Errorf(`getUsername(%v) = (%v, %v) want (%v, %v)`, test.auth, result, err, test.want, test.wantErr)
		}
	}
}
