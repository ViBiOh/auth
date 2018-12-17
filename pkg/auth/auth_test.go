package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/auth/pkg/ident"
)

func authTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Authorization`) == `` {
			http.Error(w, ident.ErrEmptyAuth.Error(), http.StatusUnauthorized)
		} else if r.Header.Get(`Authorization`) == `unauthorized` {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.Write([]byte(r.Header.Get(`Authorization`)))
		}
	}))
}

func Test_loadUsersProfiles(t *testing.T) {
	var cases = []struct {
		intention        string
		usersAndProfiles string
		want             int
	}{
		{
			`should handle empty string`,
			``,
			0,
		},
		{
			`should handle one user`,
			`admin:admin`,
			1,
		},
		{
			`should handle multiples users`,
			`admin:admin|multi,guest:,visitor:visitor`,
			3,
		},
	}

	for _, testCase := range cases {
		if result := len(loadUsersProfiles(testCase.usersAndProfiles)); result != testCase.want {
			t.Errorf("%s\nloadUsersProfiles(%+v) = %+v, want %+v", testCase.intention, testCase.usersAndProfiles, result, testCase.want)
		}
	}
}
