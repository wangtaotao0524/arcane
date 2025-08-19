package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSplitImageReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ref       string
		wantReg   string
		wantRepo  string
		wantTag   string
		wantError bool
	}{
		{
			name:     "docker hub short name adds library and latest",
			ref:      "nginx",
			wantReg:  DEFAULT_REGISTRY,
			wantRepo: "library/nginx",
			wantTag:  "latest",
		},
		{
			name:     "docker hub explicit tag",
			ref:      "nginx:alpine",
			wantReg:  DEFAULT_REGISTRY,
			wantRepo: "library/nginx",
			wantTag:  "alpine",
		},
		{
			name:     "docker.io prefix normalizes to default",
			ref:      "docker.io/library/redis:7",
			wantReg:  DEFAULT_REGISTRY,
			wantRepo: "library/redis",
			wantTag:  "7",
		},
		{
			name:     "ghcr with org/repo:tag",
			ref:      "ghcr.io/acme/api:1.2.3",
			wantReg:  "ghcr.io",
			wantRepo: "acme/api",
			wantTag:  "1.2.3",
		},
		{
			name:     "ghcr default tag latest",
			ref:      "ghcr.io/owner/repo",
			wantReg:  "ghcr.io",
			wantRepo: "owner/repo",
			wantTag:  "latest",
		},
		{
			name:     "localhost registry with port",
			ref:      "localhost:5000/myrepo:dev",
			wantReg:  "localhost:5000",
			wantRepo: "myrepo",
			wantTag:  "dev",
		},
		{
			name:     "strip digest suffix",
			ref:      "ghcr.io/acme/api:1.2.3@sha256:deadbeef",
			wantReg:  "ghcr.io",
			wantRepo: "acme/api",
			wantTag:  "1.2.3",
		},
	}

	r := NewRegistryUtils()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			reg, repo, tag, err := r.SplitImageReference(tc.ref)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if reg != tc.wantReg || repo != tc.wantRepo || tag != tc.wantTag {
				t.Fatalf("got (%q,%q,%q), want (%q,%q,%q)", reg, repo, tag, tc.wantReg, tc.wantRepo, tc.wantTag)
			}
		})
	}
}

func TestGetRegistryURL(t *testing.T) {
	t.Parallel()

	r := NewRegistryUtils()
	if got := r.GetRegistryURL("docker.io"); got != "https://index.docker.io" {
		t.Fatalf("docker.io mapped to %q, want https://index.docker.io", got)
	}
	if got := r.GetRegistryURL(DEFAULT_REGISTRY); got != "https://index.docker.io" {
		t.Fatalf("DEFAULT_REGISTRY mapped to %q, want https://index.docker.io", got)
	}
	if got := r.GetRegistryURL("ghcr.io"); got != "https://ghcr.io" {
		t.Fatalf("ghcr.io mapped to %q, want https://ghcr.io", got)
	}
	if got := r.GetRegistryURL("https://custom.registry:8443"); got != "https://custom.registry:8443" {
		t.Fatalf("https url changed to %q", got)
	}
}

func TestExtractRegistryDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ref  string
		want string
	}{
		{"nginx:alpine", "docker.io"},
		{"docker.io/library/redis:7", "docker.io"},
		{"ghcr.io/acme/api:1.2.3", "ghcr.io"},
		{"localhost:5000/repo:dev", "localhost:5000"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.ref, func(t *testing.T) {
			got, err := ExtractRegistryDomain(tc.ref)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCheckAuthParsesRealm(t *testing.T) {
	t.Parallel()

	realm := "https://auth.example/token"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.Header().Set("WWW-Authenticate", `Bearer realm="`+realm+`",service="example.org"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	utils := NewRegistryUtils()
	// Pass full server URL so GetRegistryURL returns it unchanged
	got, err := utils.CheckAuth(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != realm {
		t.Fatalf("realm = %q, want %q", got, realm)
	}
}

func TestGetToken(t *testing.T) {
	t.Parallel()

	// Fake auth server returning a token
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify optional basic auth presence if provided
		if ah := r.Header.Get("Authorization"); ah != "" {
			if !strings.HasPrefix(ah, "Basic ") {
				t.Fatalf("expected Basic auth header, got %q", ah)
			}
			// ensure decodes without error
			if _, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ah, "Basic ")); err != nil {
				t.Fatalf("bad basic auth header: %v", err)
			}
		}

		_ = r.ParseForm()
		_ = r.Form.Get("service")
		_ = r.Form.Get("scope")

		_ = json.NewEncoder(w).Encode(map[string]string{"token": "abc123"})
	}))
	defer srv.Close()

	utils := NewRegistryUtils()
	// authURL points to our test server
	token, err := utils.GetToken(context.Background(), srv.URL, "library/hello-world", &RegistryCredentials{
		Username: "user",
		Token:    "pass",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "abc123" {
		t.Fatalf("token = %q, want abc123", token)
	}
}

func TestGetLatestDigest(t *testing.T) {
	t.Parallel()

	// Fake registry returning a digest header on HEAD
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("expected HEAD, got %s", r.Method)
		}
		if !strings.Contains(r.Header.Get("Accept"), "manifest") {
			t.Fatalf("missing Accept header for manifests")
		}
		w.Header().Set("Docker-Content-Digest", "sha256:deadbeef")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	utils := NewRegistryUtils()
	digest, err := utils.GetLatestDigest(context.Background(), srv.URL, "org/repo", "latest", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if digest != "sha256:deadbeef" {
		t.Fatalf("digest = %q, want sha256:deadbeef", digest)
	}
}

func TestGetImageTagsPagination(t *testing.T) {
	t.Parallel()

	var server *httptest.Server

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/org/repo/tags/list" && (r.URL.RawQuery == "" || r.URL.RawQuery == "page=1"):
			page2URL := server.URL + "/v2/org/repo/tags/list?page=2"
			w.Header().Set("Link", `<`+page2URL+`>; rel="next"`)
			_ = json.NewEncoder(w).Encode(map[string]any{"tags": []string{"a", "b"}})
		case r.URL.Path == "/v2/org/repo/tags/list" && r.URL.RawQuery == "page=2":
			_ = json.NewEncoder(w).Encode(map[string]any{"tags": []string{"c"}})
		default:
			http.NotFound(w, r)
		}
	})

	server = httptest.NewServer(handler)
	defer server.Close()

	utils := NewRegistryUtils()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tags, err := utils.GetImageTags(ctx, server.URL, "org/repo", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.Join(tags, ",")
	if got != "a,b,c" {
		t.Fatalf("tags = %q, want %q", got, "a,b,c")
	}
}
