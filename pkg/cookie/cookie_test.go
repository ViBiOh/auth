package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v2/pkg/request"
)

func TestGetCookieValue(t *testing.T) {
	reqWithCookie := httptest.NewRequest(http.MethodGet, "/", nil)
	reqWithCookie.AddCookie(&http.Cookie{
		Name:  "Testcookie",
		Value: "cookie_content",
	})

	var cases = []struct {
		intention string
		request   *http.Request
		want      string
	}{
		{
			"should return empty value when no cookie",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"",
		},
		{
			"should return existing cookie",
			reqWithCookie,
			"cookie_content",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if result := GetCookieValue(testCase.request, "Testcookie"); result != testCase.want {
				t.Errorf("GetCookieValue(%#v) = %#v, want %#v", testCase.request, result, testCase.want)
			}
		})
	}
}

func TestSetCookieAndRedirect(t *testing.T) {
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
			"should redirect with cookie of given content",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"/redirect_with_cookie",
			"vibioh.fr",
			"secret_token",
			`<a href="/redirect_with_cookie">Found</a>.

`,
			http.StatusFound,
			"auth=secret_token; Path=/; Domain=vibioh.fr; HttpOnly; Secure",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			SetCookieAndRedirect(writer, testCase.request, testCase.redirect, testCase.cookieDomain, testCase.cookieContent)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("SetCookieAndRedirect(%#v) = %#v, want status %#v", testCase.request, result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("SetCookieAndRedirect(%#v) = %#v, want %#v", testCase.request, string(result), testCase.want)
			}

			if result := writer.Header().Get("Set-Cookie"); result != testCase.wantCookie {
				t.Errorf("SetCookieAndRedirect(%#v) = %#v, want %#v", testCase.request, result, testCase.wantCookie)
			}
		})
	}
}

func TestClearCookieAndRedirect(t *testing.T) {
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
			"should redirect with cookie of given content",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"/redirect_with_cookie",
			"vibioh.fr",
			`<a href="/redirect_with_cookie">Found</a>.

`,
			http.StatusFound,
			"auth=none; Path=/; Domain=vibioh.fr; Max-Age=0; HttpOnly; Secure",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			ClearCookieAndRedirect(writer, testCase.request, testCase.redirect, testCase.cookieDomain)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("ClearCookieAndRedirect(%#v) = %#v, want status %#v", testCase.request, result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("ClearCookieAndRedirect(%#v) = %#v, want %#v", testCase.request, string(result), testCase.want)
			}

			if result := writer.Header().Get("Set-Cookie"); result != testCase.wantCookie {
				t.Errorf("ClearCookieAndRedirect(%#v) = %#v, want %#v", testCase.request, result, testCase.wantCookie)
			}
		})
	}
}
