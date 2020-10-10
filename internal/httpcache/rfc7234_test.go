package httpcache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRFC7234(t *testing.T) {
	t.Run("store", func(t *testing.T) {
		var now = time.Now().UTC()
		var cases = []struct {
			title string
			resp  *http.Response
			store bool
		}{
			{
				title: "GET",
				resp: &http.Response{
					StatusCode: 200,
					Header: http.Header{
						"Cache-Control": []string{"max-age=5"},
					},
					Request: &http.Request{
						Method: "GET",
					},
				},
				store: true,
			},
			{
				title: "HEAD",
				resp: &http.Response{
					StatusCode: 200,
					Header: http.Header{
						"Cache-Control": []string{"max-age=5"},
					},
					Request: &http.Request{
						Method: "HEAD",
					},
				},
				store: true,
			},
			{
				title: "POST",
				resp: &http.Response{
					StatusCode: 200,
					Request: &http.Request{
						Method: "POST",
					},
				},
				store: false,
			},
			{
				title: "GET 500",
				resp: &http.Response{
					StatusCode: 500,
					Request: &http.Request{
						Method: "GET",
					},
				},
				store: false,
			},
			{
				title: "GET request no-store",
				resp: &http.Response{
					StatusCode: 200,
					Request: &http.Request{
						Method: "GET",
						Header: http.Header{
							"Cache-Control": []string{"no-store"},
						},
					},
				},
				store: false,
			},
			{
				title: "GET response no-store",
				resp: &http.Response{
					StatusCode: 200,
					Header: http.Header{
						"Cache-Control": []string{"no-store"},
					},
					Request: &http.Request{
						Method: "GET",
					},
				},
				store: false,
			},
			{
				title: "GET response expired",
				resp: &http.Response{
					StatusCode: 200,
					Header: http.Header{
						"Date":    []string{now.Format(time.RFC1123)},
						"Expires": []string{now.Add(-time.Minute).Format(time.RFC1123)},
					},
					Request: &http.Request{
						Method: "GET",
					},
				},
				store: false,
			},
			{
				title: "GET response expires",
				resp: &http.Response{
					StatusCode: 200,
					Header: http.Header{
						"Date":    []string{now.Format(time.RFC1123)},
						"Expires": []string{now.Add(time.Minute).Format(time.RFC1123)},
					},
					Request: &http.Request{
						Method: "GET",
					},
				},
				store: true,
			},
			{
				title: "GET no explicit cache",
				resp: &http.Response{
					StatusCode: 200,
					Header:     http.Header{},
					Request: &http.Request{
						Method: "GET",
					},
				},
				store: false,
			},
		}

		for _, c := range cases {
			t.Run(c.title, func(t *testing.T) {
				var assert = require.New(t)
				var strategy = RFC7234{}

				assert.Equal(c.store, strategy.store(c.resp))
			})
		}
	})

	t.Run("fresh", func(t *testing.T) {
		var now = time.Now()
		var cases = []struct {
			title string
			resp  *http.Response
			fresh Freshness
		}{
			{
				title: "no-cache request",
				resp: &http.Response{
					Request: &http.Request{
						Header: http.Header{
							"Cache-Control": []string{"no-cache"},
						},
					},
				},
				fresh: Stale,
			},
			{
				title: "no-cache response",
				resp: &http.Response{
					Header: http.Header{
						"Cache-Control": []string{"no-cache"},
					},
					Request: &http.Request{},
				},
				fresh: Stale,
			},
			{
				title: "vary mismatch",
				resp: &http.Response{
					Request: &http.Request{
						Header: http.Header{
							"Vary":            []string{"Accept-Language"},
							"Accept-Language": []string{"en-US"},
						},
					},
				},
				fresh: Transparent,
			},
			{
				title: "fresh",
				resp: &http.Response{
					Request: &http.Request{},
					Header: http.Header{
						"Date":          []string{now.Format(time.RFC1123)},
						"Cache-Control": []string{"max-age=5"},
					},
				},
				fresh: Fresh,
			},
			{
				title: "stale",
				resp: &http.Response{
					Request: &http.Request{},
				},
				fresh: Stale,
			},
		}

		for _, c := range cases {
			t.Run(c.title, func(t *testing.T) {
				var assert = require.New(t)
				var strategy = RFC7234{}

				assert.Equal(c.fresh, strategy.fresh(c.resp))
			})
		}
	})
}
