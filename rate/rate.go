package rate

import (
	"net/http"
	"time"
)

const ipRateDelay = time.Second * -60
const ipRateCount = 60

type rateLimit struct {
	ip    string
	calls []time.Time
}

var userRate = make(map[string]*rateLimit, 0)

// CheckRate verify that request respect rate limit
func CheckRate(r *http.Request) bool {
	ip := r.RemoteAddr
	rate, ok := userRate[ip]

	if !ok {
		rate = &rateLimit{ip, make([]time.Time, 0)}
		userRate[ip] = rate
	}

	rate.calls = append(rate.calls, time.Now())

	nowMinusDelay := time.Now().Add(ipRateDelay)
	for len(rate.calls) > 0 && rate.calls[0].Before(nowMinusDelay) {
		rate.calls = rate.calls[1:]
	}

	return len(rate.calls) >= ipRateCount
}
