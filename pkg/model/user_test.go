package model

import (
	"reflect"
	"testing"
)

func TestNewUser(t *testing.T) {
	var cases = []struct {
		intention string
		id        uint64
		username  string
		email     string
		profiles  string
		want      *User
	}{
		{
			"should work with given params",
			1,
			"vibioh",
			"nobody@localhost",
			"admin|multi",
			&User{1, "vibioh", "nobody@localhost", "admin|multi"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := NewUser(testCase.id, testCase.username, testCase.email, testCase.profiles); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("NewUser(%#v, %#v, %#v) = %#v, want %#v", testCase.id, testCase.username, testCase.profiles, result, testCase.want)
			}
		})
	}
}

func TestHasProfile(t *testing.T) {
	var cases = []struct {
		intention string
		instance  User
		profile   string
		want      bool
	}{
		{
			"should handle nil profiles",
			User{},
			"admin",
			false,
		},
		{
			"should find simple match",
			User{profiles: "admin"},
			"admin",
			true,
		},
		{
			"should find match when multiples values",
			User{profiles: "admin|multi"},
			"multi",
			true,
		},
		{
			"should find no match",
			User{profiles: "multi"},
			"admin",
			false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := testCase.instance.HasProfile(testCase.profile); result != testCase.want {
				t.Errorf("%#v.HasProfile(%#v) = %#v, want %#v", testCase.instance, testCase.profile, result, testCase.want)
			}
		})
	}
}
