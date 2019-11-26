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
		profiles  string
		want      User
	}{
		{
			"should work with given params",
			1,
			"vibioh",
			"admin|multi",
			User{ID: 1, Username: "vibioh", profiles: "admin|multi"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := NewUser(testCase.id, testCase.username, testCase.profiles); !reflect.DeepEqual(result, testCase.want) {
				t.Errorf("NewUser() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestSetID(t *testing.T) {
	var cases = []struct {
		intention string
		instance  User
		input     uint64
	}{
		{
			"simple",
			NewUser(0, "test", ""),
			8000,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			testCase.instance.SetID(testCase.input)
			if testCase.instance.ID != testCase.input {
				t.Errorf("SetID() = %d, want %d", testCase.instance.ID, testCase.input)
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
				t.Errorf("HasProfile() = %t, want %t", result, testCase.want)
			}
		})
	}
}
