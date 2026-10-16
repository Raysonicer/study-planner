package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

var testFS = fstest.MapFS{
	"index.html":    {Data: []byte("<html>spa-index</html>")},
	"assets/app.js": {Data: []byte("console.log('app')")},
}

func TestSPAIndex(t *testing.T) {
	h := SPAHandler(testFS)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "<html>spa-index</html>" {
		t.Fatalf("body = %q", got)
	}
}

func TestSPAFallbackToIndex(t *testing.T) {
	h := SPAHandler(testFS)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/week", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	if rec.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("missing no-cache header for index fallback")
	}
}

func TestSPAStaticAsset(t *testing.T) {
	h := SPAHandler(testFS)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "console.log('app')" {
		t.Fatalf("body = %q", body)
	}
}
