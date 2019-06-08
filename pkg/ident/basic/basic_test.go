package basic

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ViBiOh/auth/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

func TestLoadUsers(t *testing.T) {
	var cases = []struct {
		intention string
		input     string
		want      int
		wantErr   error
	}{
		{
			"should handle empty string",
			"",
			0,
			nil,
		},
		{
			"should handle invalid format",
			"invalid_username",
			0,
			errors.New("invalid format of user for invalid_username"),
		},
		{
			"should handle valid format",
			"anc:admin:admin,1:guest:guest",
			2,
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			users, err := loadUsers(testCase.input)
			result := len(users)

			failed := false

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
				t.Errorf("LoadUsers(%#v) = (%#v, %#v), want (%#v, %#v)", testCase.input, result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	var cases = []struct {
		intention string
		users     string
		want      int
		wantErr   error
	}{
		{
			"should handle load error",
			"invalid format",
			0,
			errors.New("invalid format of user for invalid format"),
		},
		{
			"should load users from given args",
			"1:admin:admin",
			1,
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			auth, err := New(Config{users: &testCase.users}, nil)
			var authClient *App
			if auth != nil {
				authClient = auth.(*App)
			}

			failed := false

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
				t.Errorf("New(%#v) = (%#v, %#v), want (%#v, %#v)", testCase.users, authClient.users, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestGetName(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"should return constant",
			"Basic",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := (&App{}).GetName(); result != testCase.want {
				t.Errorf("GetName() = %#v, want %#v", result, testCase.want)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	password, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
	authClient := App{}
	authClient.users = map[string]*basicUser{"admin": {model.NewUser("0", "admin", "", ""), password}}

	var cases = []struct {
		intention string
		auth      string
		want      *model.User
		wantErr   error
	}{
		{
			"should handle malformed header",
			"admin",
			nil,
			errors.New("illegal base64 data at input byte 4"),
		},
		{
			"should handle malformed content",
			base64.StdEncoding.EncodeToString([]byte("AdMiN")),
			nil,
			errors.New("invalid format for basic auth"),
		},
		{
			"should handle not found user",
			base64.StdEncoding.EncodeToString([]byte("guest:password")),
			nil,
			errors.New("invalid credentials"),
		},
		{
			"should handle invalid credentials",
			base64.StdEncoding.EncodeToString([]byte("AdMiN:admin")),
			nil,
			errors.New("invalid credentials"),
		},
		{
			"should handle valid auth",
			base64.StdEncoding.EncodeToString([]byte("AdMiN:password")),
			&model.User{ID: "0", Username: "admin"},
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := authClient.GetUser(nil, testCase.auth)

			failed := false

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
				t.Errorf("GetUser(%#v) = (%#v, %#v) want (%#v, %#v)", testCase.auth, result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestRedirect(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"should return Basic redirection",
			"/login/basic",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			(&App{}).Redirect(writer, httptest.NewRequest(http.MethodGet, "/", nil))
			result := writer.Header().Get("location")
			if result != testCase.want {
				t.Errorf("Redirect() = (%#v), want (%#v)", result, testCase.want)
			}
		})
	}
}
