package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils"
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
			`should return empty value when no cookie`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			``,
		},
		{
			`should return existing cookie`,
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

func Test_SetCookieAndRedirect(t *testing.T) {
	var cases = []struct {
		intention     string
		request       *http.Request
		redirect      string
		cookieDomain  string
		cookieContent string
		want          string
		wantStatus    int
		wantCookie    string
	}{
		{
			`should redirect with cookie of given content`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			`/redirect_with_cookie`,
			`vibioh.fr`,
			`secret_token`,
			`<a href="/redirect_with_cookie">Found</a>.

`,
			http.StatusFound,
			`auth=secret_token; Path=/; Domain=vibioh.fr; HttpOnly; Secure`,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()

		SetCookieAndRedirect(writer, testCase.request, testCase.redirect, testCase.cookieDomain, testCase.cookieContent)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%s\nSetCookieAndRedirect(%+v) = %+v, want status %+v", testCase.intention, testCase.request, result, testCase.wantStatus)
		}

		if result, _ := httputils.ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf("%s\nSetCookieAndRedirect(%+v) = %+v, want %+v", testCase.intention, testCase.request, string(result), testCase.want)
		}

		if result := writer.Header().Get(`Set-Cookie`); result != testCase.wantCookie {
			t.Errorf("%s\nSetCookieAndRedirect(%+v) = %+v, want %+v", testCase.intention, testCase.request, result, testCase.wantCookie)
		}
	}
}

func Test_ClearCookieAndRedirect(t *testing.T) {
	var cases = []struct {
		intention    string
		request      *http.Request
		redirect     string
		cookieDomain string
		want         string
		wantStatus   int
		wantCookie   string
	}{
		{
			`should redirect with cookie of given content`,
			httptest.NewRequest(http.MethodGet, `/`, nil),
			`/redirect_with_cookie`,
			`vibioh.fr`,
			`<a href="/redirect_with_cookie">Found</a>.

`,
			http.StatusFound,
			`auth=none; Path=/; Domain=vibioh.fr; Max-Age=0; HttpOnly; Secure`,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()

		ClearCookieAndRedirect(writer, testCase.request, testCase.redirect, testCase.cookieDomain)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%s\nClearCookieAndRedirect(%+v) = %+v, want status %+v", testCase.intention, testCase.request, result, testCase.wantStatus)
		}

		if result, _ := httputils.ReadBody(writer.Result().Body); string(result) != testCase.want {
			t.Errorf("%s\nClearCookieAndRedirect(%+v) = %+v, want %+v", testCase.intention, testCase.request, string(result), testCase.want)
		}

		if result := writer.Header().Get(`Set-Cookie`); result != testCase.wantCookie {
			t.Errorf("%s\nClearCookieAndRedirect(%+v) = %+v, want %+v", testCase.intention, testCase.request, result, testCase.wantCookie)
		}
	}
}
