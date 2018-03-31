package provider

import (
	"reflect"
	"testing"
)

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
