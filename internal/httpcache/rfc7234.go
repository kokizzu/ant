package httpcache

import (
	"net/http"
)

// RFC7234 implements the standard cache strategy.
//
// https://tools.ietf.org/html/rfc7234
type RFC7234 struct{}

// Cache implementation.
//
// The method returns true if the request may use a cached
// response, or if it allows caching.
func (RFC7234) cache(req *http.Request) bool {
	return (req.Method == "GET" || req.Method == "HEAD") && !nostore(req.Header)
}

// Store implementation.
//
// https://tools.ietf.org/html/rfc7234#section-3
func (RFC7234) store(resp *http.Response) bool {
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

	// the "no-store" cache directive (see Section 5.2) does not appear
	// in request or response header fields.
	if nostore(req.Header) || nostore(resp.Header) {
		return false
	}

	// The response has an explicit "lifetime" duration.
	age, ok := lifetime(resp)
	return ok && age > 0
}

// Fresh implementation.
//
// https://tools.ietf.org/html/rfc7234#section-4
func (RFC7234) fresh(resp *http.Response) Freshness {
	var req = resp.Request

	// selecting header fields nominated by the stored response (if any)
	// match those presented (see Section 4.1).
	if !matches(req, resp) {
		return Transparent
	}

	// Parse request and response directives.
	var (
		reqd = directivesFrom(req.Header)
		resd = directivesFrom(resp.Header)
	)

	// the presented request does not contain the no-cache pragma
	// (Section 5.4), nor the no-cache cache directive (Section 5.2.1),
	// unless the stored response is successfully validated (Section 4.3).
	//
	// the stored response does not contain the no-cache cache directive
	// (Section 5.2.2.2), unless it is successfully validated (Section 4.3)
	if reqd.has("no-cache") || resd.has("no-cache") {
		return Stale
	}

	// the stored response is either fresh (see Section 4.2).
	if age, ok := lifetime(resp); ok && age > 0 {
		return Fresh
	}

	// validate (see Section 4.3).
	return Stale
}
