package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_GetCookieValue(t *testing.T) {
	reqWithCookie := httptest.NewRequest(http.MethodGet, `/`, nil)
	reqWithCookie.AddCookie(&http.Cookie{
		Name:  `test_cookie`,
		Value: `cookie_content`,
	})

	var cases = []struct {
		intention string
		request   *http.Request
		want      string
	}{
		{
			`should work with given params`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			``,
		},
		{
			`should work with given params`,
			reqWithCookie,
			`cookie_content`,
		},
	}

	for _, testCase := range cases {
		if result := GetCookieValue(testCase.request, `test_cookie`); result != testCase.want {
			t.Errorf("%s\nGetCookieValue(%+v) = %+v, want %+v", testCase.intention, testCase.request, result, testCase.want)
		}
	}
}
