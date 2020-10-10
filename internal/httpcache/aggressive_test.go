package httpcache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAggressive(t *testing.T) {
	t.Run("store", func(t *testing.T) {
		// var now = time.Now().UTC()
		var cases = []struct {
			title string
			resp  *http.Response
			store bool
		}{
			{
				title: "GET",
				resp: &http.Response{
					StatusCode: 200,
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
		}

		for _, c := range cases {
			t.Run(c.title, func(t *testing.T) {
				var assert = require.New(t)
				var strategy = Aggressive{}

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
				title: "fresh",
				resp: &http.Response{
					Request: &http.Request{},
					Header: http.Header{
						"Date": []string{now.Format(time.RFC1123)},
					},
				},
				fresh: Fresh,
			},
			{
				title: "fresh 2 hours",
				resp: &http.Response{
					Request: &http.Request{},
					Header: http.Header{
						"Date": []string{now.Add(-(2 * time.Hour)).Format(time.RFC1123)},
					},
				},
				fresh: Fresh,
			},
			{
				title: "transparent 2 days",
				resp: &http.Response{
					Request: &http.Request{},
					Header: http.Header{
						"Date": []string{now.Add(-(48 * time.Hour)).Format(time.RFC1123)},
					},
				},
				fresh: Transparent,
			},
			{
				title: "transparent",
				resp: &http.Response{
					Request: &http.Request{},
				},
				fresh: Transparent,
			},
		}

		for _, c := range cases {
			t.Run(c.title, func(t *testing.T) {
				var assert = require.New(t)
				var strategy = Aggressive{}

				assert.Equal(c.fresh, strategy.fresh(c.resp))
			})
		}
	})
}
