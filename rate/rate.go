package rate

import (
	"flag"
	"net/http"
	"time"
)

var (
	ipRateDelay = flag.Duration(`rateDelay`, time.Second*60, `Rate IP delay`)
	ipRateCount = flag.Int(`rateCount`, 60, `Rate IP count`)
)

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

	now := time.Now()
	rate.calls = append(rate.calls, now)

	nowMinusDelay := now.Add(*ipRateDelay * -1)
	for len(rate.calls) > 0 && rate.calls[0].Before(nowMinusDelay) {
		rate.calls = rate.calls[1:]
	}

	return len(rate.calls) < *ipRateCount
}
