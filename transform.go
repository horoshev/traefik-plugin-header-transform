// nolint
package traefik_plugin_header_transform

import (
	"context"
	"net/http"
	"strings"
)

const (
	Cookie = "@Cookie:"
	Header = "@Header:"
)

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// Config holds the plugin configuration.
type Config struct {
	Transformers []Transform `json:"rewrites,omitempty"`
}

// Rewrite holds one rewrite body configuration.
type Transform struct {
	Header string `json:"header,omitempty"`
	Value  string `json:"value,omitempty"`
}

type middleware struct {
	name         string
	next         http.Handler
	transformers []transformer
}

type transformer struct {
	header string
	fn     transformFunc
}

// New creates and returns a new rewrite body plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	transformers := make([]transformer, 0, len(config.Transformers))

	for _, t := range config.Transformers {
		transformers = append(transformers, transformer{
			header: t.Header,
			fn:     NewTransformer(t.Value),
		})
	}

	return &middleware{
		name:         name,
		next:         next,
		transformers: transformers,
	}, nil
}

func (r *middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, transform := range r.transformers {
		req.Header.Set(transform.header, transform.fn(req))
	}

	r.next.ServeHTTP(rw, req)
}

type transformFunc func(req *http.Request) string

func NewTransformer(value string) transformFunc {
	if cookie, ok := strings.CutPrefix(value, Cookie); ok {
		return CookieTransformer(cookie)
	}

	if header, ok := strings.CutPrefix(value, Header); ok {
		return HeaderTransformer(header)
	}

	return ExactTransformer(value)
}

func CookieTransformer(name string) transformFunc {
	return func(req *http.Request) string {
		cookie, err := req.Cookie(name)
		if err != nil {
			return ""
		}

		return cookie.Value
	}
}

func HeaderTransformer(header string) transformFunc {
	return func(req *http.Request) string {
		return strings.Join(req.Header.Values(header), ",")
	}
}

func ExactTransformer(value string) transformFunc {
	return func(req *http.Request) string {
		return value
	}
}
