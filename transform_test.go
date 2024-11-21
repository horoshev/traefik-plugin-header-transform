package traefik_plugin_header_transform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		desc         string
		transformers []Transform
		reqHeader    http.Header
		expReqHeader http.Header
	}{
		{
			desc: "should create http header X-Auth from cookie Authorization",
			transformers: []Transform{
				{
					Header: "X-Auth",
					Value:  "@Cookie:Authorization",
				},
			},
			reqHeader: map[string][]string{
				"Cookie": {
					"foo",
					"Authorization=abc",
				},
			},
			expReqHeader: map[string][]string{
				"Cookie": {
					"foo",
					"Authorization=abc",
				},
				"X-Auth": {"abc"},
			},
		},
		{
			desc: "should create http header X-Forwarded-Host from header Host",
			transformers: []Transform{
				{
					Header: "X-Forwarded-Host",
					Value:  "@Header:Host",
				},
			},
			reqHeader: map[string][]string{
				"Host": {"test:1000"},
			},
			expReqHeader: map[string][]string{
				"Host":             {"test:1000"},
				"X-Forwarded-Host": {"test:1000"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{
				Transformers: test.transformers,
			}

			next := func(rw http.ResponseWriter, req *http.Request) {
				for k, v := range req.Header {
					for _, h := range v {
						rw.Header().Add(k, h)
					}
				}

				rw.WriteHeader(http.StatusOK)
			}

			rewriteBody, err := New(context.Background(), http.HandlerFunc(next), config, "rewriteHeader")
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			for k, v := range test.reqHeader {
				for _, h := range v {
					req.Header.Add(k, h)
				}
			}

			rewriteBody.ServeHTTP(recorder, req)
			for k, expected := range test.expReqHeader {
				values := recorder.Header().Values(k)

				assert.Equal(t, values, expected, "slices arent equals")
			}
		})
	}
}
