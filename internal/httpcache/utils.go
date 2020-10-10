package httpcache

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Matches ensures that the given request and response match.
//
// https://tools.ietf.org/html/rfc7234#section-4.1
func matches(req *http.Request, resp *http.Response) bool {
	var vary = req.Header.Get("Vary")

	for _, h := range split(vary, ",") {
		if key := http.CanonicalHeaderKey(h); key != "" {
			if req.Header.Get(key) != resp.Header.Get(key) {
				return false
			}
		}
	}

	return true
}

// Nocache returns true if the no-cache is set.
func nocache(h http.Header) bool {
	var c = h.Get("Cache-Control")
	var p = h.Get("Pragma")

	for _, v := range split(c, ",") {
		if v == "no-cache" {
			return true
		}
	}

	for _, v := range split(p, ",") {
		if v == "no-cache" {
			return true
		}
	}

	return false
}

// Nostore returns true if no-store is set.
func nostore(h http.Header) bool {
	var c = h.Get("Cache-Control")

	for _, v := range split(c, ",") {
		if v == "no-store" {
			return true
		}
	}

	return false
}

// Maxage returns the cache-control max-age.
//
// When max-age does not exist, ok is false.
func maxage(h http.Header) (age time.Duration, ok bool) {
	var c = h.Get("Cache-Control")

	for _, d := range split(c, ",") {
		if strings.HasPrefix(d, "max-age") {
			if j := strings.IndexByte(d, '='); j != -1 {
				n, _ := strconv.ParseInt(d[j+1:], 10, 64)
				age, ok = time.Duration(n)*time.Second, true
				break
			}
		}
	}

	return
}

// Expires returns the expires timestamp.
//
// When expires does not exist or is zero, ok is false.
func expires(h http.Header) (expires time.Time, ok bool) {
	if v := h.Get("Expires"); v != "" {
		t, err := time.Parse(time.RFC1123, v)
		expires, ok = t, (err == nil && !t.IsZero())
	}
	return
}

// Date returns the date timestamp.
//
// When date does not exist or is zero, ok is false.
func date(h http.Header) (date time.Time, ok bool) {
	if v := h.Get("Date"); v != "" {
		t, err := time.Parse(time.RFC1123, v)
		date, ok = t, (err == nil && !t.IsZero())
	}
	return
}

// Split splits the given str by sep.
//
// The method omits any empty values and normalizes
// the values by lowercasing them.
func split(str, sep string) (ret []string) {
	for _, v := range strings.Split(str, sep) {
		if v := strings.TrimSpace(v); v != "" {
			ret = append(ret, strings.ToLower(v))
		}
	}
	return
}

// Lifetime returns the lifetime duration of the response.
//
// https://tools.ietf.org/html/rfc7234#section-4.2.1
func lifetime(resp *http.Response) (time.Duration, bool) {
	if age, ok := maxage(resp.Header); ok {
		return age, true
	}

	if exp, ok := expires(resp.Header); ok {
		if date, ok := date(resp.Header); ok {
			return exp.Sub(date), true
		}
	}

	return -1, false
}
