package httpcache

import (
	"net/http"
	"time"
)

// Aggressive represents an aggressive caching strategy.
//
// The strategy attempts to cache all HEAD and GET requests
// regardless of cache-control directives up to a configurable
// duration with the default of 1 day.
type Aggressive struct {
	// The lifetime duration for all responses.
	//
	// If <= 0, defaults to 1 day.
	Lifetime time.Duration
}

// Cache implementation.
func (a Aggressive) cache(req *http.Request) bool {
	return req.Method == "GET" || req.Method == "HEAD"
}

// Store implementation.
func (a Aggressive) store(resp *http.Response) bool {
	var req = resp.Request

	// The request method is cacheable.
	switch req.Method {
	case "GET":
	case "HEAD":
	default:
		return false
	}

	// The response status code is cacheable.
	switch resp.StatusCode {
	case 200, 203, 204, 206:
	case 300, 301:
	case 404, 405, 410, 414:
	case 501:
	default:
		return false
	}

	return true
}

// Fresh implementation.
func (a Aggressive) fresh(resp *http.Response) Freshness {
	if date, ok := date(resp.Header); ok {
		if time.Since(date) < a.lifetime() {
			return Fresh
		}
	}
	return Transparent
}

// Lifetime returns the lifetime.
func (a Aggressive) lifetime() time.Duration {
	if a.Lifetime > 0 {
		return a.Lifetime
	}
	return 24 * time.Hour
}
