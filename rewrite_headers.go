// nolint
package traefik_plugin_rewrite_headers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
)

// Rewrite holds one rewrite body configuration.
type Rewrite struct {
	Header      string `json:"header,omitempty"`
	Regex       string `json:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty"`
}

// Config holds the plugin configuration.
type Config struct {
	Rewrites []Rewrite `json:"rewrites,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type rewrite struct {
	header      string
	regex       *regexp.Regexp
	replacement string
}

type rewriteBody struct {
	name     string
	next     http.Handler
	rewrites []rewrite
}

// New creates and returns a new rewrite body plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if next == nil {
		return nil, fmt.Errorf("next handler cannot be nil")
	}

	rewrites := make([]rewrite, 0, len(config.Rewrites))

	for i, rewriteConfig := range config.Rewrites {
		if rewriteConfig.Header == "" {
			return nil, fmt.Errorf("rewrite %d missing header", i)
		}
		if rewriteConfig.Regex == "" {
			return nil, fmt.Errorf("rewrite %d missing regex", i)
		}

		regex, err := regexp.Compile(rewriteConfig.Regex)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", rewriteConfig.Regex, err)
		}

		rewrites = append(rewrites, rewrite{
			header:      rewriteConfig.Header,
			regex:       regex,
			replacement: rewriteConfig.Replacement,
		})
	}

	return &rewriteBody{
		name:     name,
		next:     next,
		rewrites: rewrites,
	}, nil
}

func (r *rewriteBody) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wrappedWriter := &responseWriter{
		writer:   rw,
		rewrites: r.rewrites,
	}

	r.next.ServeHTTP(wrappedWriter, req)
}

type responseWriter struct {
	writer      http.ResponseWriter
	rewrites    []rewrite
	wroteHeader bool
}

func (r *responseWriter) Header() http.Header {
	return r.writer.Header()
}

func (r *responseWriter) Write(bytes []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	return r.writer.Write(bytes)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}

	r.wroteHeader = true

	for _, rewrite := range r.rewrites {
		headers := r.writer.Header().Values(rewrite.header)

		if len(headers) == 0 {
			continue
		}

		r.writer.Header().Del(rewrite.header)

		for _, header := range headers {
			value := rewrite.regex.ReplaceAllString(header, rewrite.replacement)
			r.writer.Header().Add(rewrite.header, value)
		}
	}

	r.writer.WriteHeader(statusCode)
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.writer.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not a http.Hijacker", r.writer)
	}

	return hijacker.Hijack()
}

func (r *responseWriter) Flush() {
	if flusher, ok := r.writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := r.writer.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}

	return http.ErrNotSupported
}

func (r *responseWriter) ReadFrom(src io.Reader) (int64, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	if rf, ok := r.writer.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}

	return io.Copy(r.writer, src)
}
