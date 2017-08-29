package rate

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckRate(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.RemoteAddr = `localhost`

	calls := make([]time.Time, *ipRateCount)
	for i := 0; i < *ipRateCount; i++ {
		calls[i] = time.Now()
	}

	var tests = []struct {
		userRate map[string]*rateLimit
		want     bool
	}{
		{
			map[string]*rateLimit{},
			true,
		},
		{
			map[string]*rateLimit{
				`localhost`: {
					ip: `localhost`,
					calls: []time.Time{
						time.Now(),
					},
				},
			},
			true,
		},
		{
			map[string]*rateLimit{
				`localhost`: {
					ip: `localhost`,
					calls: []time.Time{
						time.Now().Add(-180 * time.Second),
						time.Now().Add(-90 * time.Second),
						time.Now().Add(-60 * time.Second),
						time.Now().Add(-30 * time.Second),
					},
				},
			},
			true,
		},
		{
			map[string]*rateLimit{
				`localhost`: {
					ip:    `localhost`,
					calls: calls,
				},
			},
			false,
		},
	}

	for _, test := range tests {
		userRate = test.userRate

		if result := CheckRate(request); result != test.want {
			t.Errorf(`CheckRate(%v) = (%v), want (%v)`, test.userRate, result, test.want)
		}
	}
}

func BenchmarkCheckRate(b *testing.B) {
	request := httptest.NewRequest(http.MethodGet, `/test`, nil)
	request.RemoteAddr = `localhost`

	calls := make([]time.Time, *ipRateCount)
	for i := 0; i < *ipRateCount; i++ {
		calls[i] = time.Now()
	}

	var test = struct {
		userRate map[string]*rateLimit
		want     bool
	}{
		map[string]*rateLimit{
			`localhost`: {
				ip:    `localhost`,
				calls: calls,
			},
		},
		false,
	}

	for i := 0; i < b.N; i++ {
		userRate = test.userRate

		if result := CheckRate(request); result != test.want {
			b.Errorf(`CheckRate(%v) = (%v), want (%v)`, test.userRate, result, test.want)
		}
	}
}
