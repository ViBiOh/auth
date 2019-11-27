package model

import (
	"reflect"
	"testing"
)

func TestNewUser(t *testing.T) {
	var cases = []struct {
		intention string
		id        uint64
		login     string
		want      User
	}{
		{
			"should work with given params",
			1,
			"vibioh",
			User{ID: 1, Login: "vibioh"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := NewUser(testCase.id, testCase.login); !reflect.DeepEqual(result, testCase.want) {
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
			NewUser(0, "test"),
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
