package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	h := &Handler{}
	lib, err := newTestLibrary()
	if err != nil {
		t.Fatal(err)
	}
	h.library = lib
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func post(mux *http.ServeMux, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("POST", "/api/simulate", strings.NewReader(body)))
	return rec
}

func TestSimulateSetsJSONContentType(t *testing.T) {
	rec := post(newTestMux(t), `{"particles":50,"generations":3}`)
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("got %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header")
	}
}

func TestSimulateRejectsBadInput(t *testing.T) {
	mux := newTestMux(t)
	for _, body := range []string{
		`{`, `{"particles":100000000}`, `{"generations":99999}`,
		`{"thickness":1e9}`, `{"geometry":"cube"}`, `{"isotope":"Xx-1"}`,
	} {
		rec := post(mux, body)
		if rec.Code != 400 || !strings.Contains(rec.Body.String(), `"error"`) {
			t.Errorf("%s: got %d %s", body, rec.Code, rec.Body.String())
		}
	}
}

func TestPreflight(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestMux(t).ServeHTTP(rec, httptest.NewRequest("OPTIONS", "/api/simulate", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("got %d", rec.Code)
	}
}
