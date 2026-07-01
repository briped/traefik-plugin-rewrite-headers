//nolint
package traefik_plugin_rewrite_headers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		desc          string
		rewrites      []Rewrite
		reqHeader     http.Header
		expRespHeader http.Header
	}{
		{
			desc: "should replace foo by bar in location header",
			rewrites: []Rewrite{
				{
					Header:      "Location",
					Regex:       "foo",
					Replacement: "bar",
				},
			},
			reqHeader: map[string][]string{
				"Location": {"foo", "anotherfoo"},
			},
			expRespHeader: map[string][]string{
				"Location": {"bar", "anotherbar"},
			},
		},
		{
			desc: "should replace http by https in location header",
			rewrites: []Rewrite{
				{
					Header:      "Location",
					Regex:       "^http://(.+)$",
					Replacement: "https://$1",
				},
			},
			reqHeader: map[string][]string{
				"Location": {"http://test:1000"},
			},
			expRespHeader: map[string][]string{
				"Location": {"https://test:1000"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{
				Rewrites: test.rewrites,
			}

			next := func(rw http.ResponseWriter, req *http.Request) {
				for k, v := range test.reqHeader {
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

			rewriteBody.ServeHTTP(recorder, req)
			for k, expected := range test.expRespHeader {
				values := recorder.Header().Values(k)

				if !testEq(values, expected) {
					t.Errorf("Slice arent equals: expect: %+v, result: %+v", expected, values)
				}
			}
		})
	}
}

func TestServeHTTP_ImplicitWrite(t *testing.T) {
	config := &Config{
		Rewrites: []Rewrite{
			{
				Header:      "Location",
				Regex:       "foo",
				Replacement: "bar",
			},
		},
	}

	next := func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Location", "foo")
		_, _ = rw.Write([]byte("ok"))
	}

	rewriteBody, err := New(context.Background(), http.HandlerFunc(next), config, "rewriteHeader")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rewriteBody.ServeHTTP(recorder, req)

	if got := recorder.Header().Get("Location"); got != "bar" {
		t.Fatalf("unexpected header value: got %q, want %q", got, "bar")
	}
}

func TestNew_Validation(t *testing.T) {
	_, err := New(context.Background(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), nil, "rewriteHeader")
	if err == nil {
		t.Fatal("expected error when config is nil")
	}

	_, err = New(context.Background(), nil, &Config{}, "rewriteHeader")
	if err == nil {
		t.Fatal("expected error when next handler is nil")
	}

	_, err = New(context.Background(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), &Config{Rewrites: []Rewrite{{Header: "", Regex: "foo"}}}, "rewriteHeader")
	if err == nil {
		t.Fatal("expected error when header is empty")
	}

	_, err = New(context.Background(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), &Config{Rewrites: []Rewrite{{Header: "Location", Regex: ""}}}, "rewriteHeader")
	if err == nil {
		t.Fatal("expected error when regex is empty")
	}
}

func TestNew_InvalidRegex(t *testing.T) {
	_, err := New(context.Background(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), &Config{Rewrites: []Rewrite{{Header: "Location", Regex: "["}}}, "rewriteHeader")
	if err == nil {
		t.Fatal("expected error when regex fails to compile")
	}
}

func testEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
