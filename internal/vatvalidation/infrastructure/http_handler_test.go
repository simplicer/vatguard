package infrastructure

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antonio-villamarin/vat-validator/internal/vatvalidation/application"
)

func TestHandleVAT_Success_EU(t *testing.T) {
	vies := &stubVIES{
		result: &application.VIESResult{
			Valid:       true,
			CompanyName: "TEST SL",
			Address:     "ADDR",
		},
	}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vat/ES00000000T", nil)
	w := httptest.NewRecorder()

	handler.HandleVAT(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp ValidationResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	// VIES should NOT be called for /vat endpoint
	data := resp.Data.(map[string]interface{})
	if data["vies_registered"] != nil {
		t.Error("VIESRegistered should be nil on /vat endpoint")
	}
}

func TestHandleVIES_Success_EU(t *testing.T) {
	vies := &stubVIES{
		result: &application.VIESResult{
			Valid:       true,
			CompanyName: "TEST SL",
			Address:     "ADDR",
		},
	}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vies/ES00000000T", nil)
	w := httptest.NewRecorder()

	handler.HandleVIES(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp ValidationResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	data := resp.Data.(map[string]interface{})
	if data["vies_registered"] == nil {
		t.Error("VIESRegistered should NOT be nil on /vies endpoint")
	}
}

func TestHandleVAT_SanctionedCountry(t *testing.T) {
	vies := &stubVIES{}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vat/RU1234567890", nil)
	w := httptest.NewRecorder()

	handler.HandleVAT(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}

	var prob ProblemJSON
	if err := json.NewDecoder(w.Body).Decode(&prob); err != nil {
		t.Fatalf("decoding problem: %v", err)
	}
	if prob.Type != "https://vat-validator.dev/errors/sanctioned-country" {
		t.Errorf("problem type = %q, want sanctioned-country", prob.Type)
	}
}

func TestHandleVAT_Unparseable(t *testing.T) {
	vies := &stubVIES{}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vat/12", nil)
	w := httptest.NewRecorder()

	handler.HandleVAT(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

func TestHandleVIES_Unavailable(t *testing.T) {
	vies := &stubVIES{err: errors.New("unavailable")}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vies/DE123456789", nil)
	w := httptest.NewRecorder()

	handler.HandleVIES(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleVAT_NonEU(t *testing.T) {
	vies := &stubVIES{}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vat/GB123456789", nil)
	w := httptest.NewRecorder()

	handler.HandleVAT(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleHealthz(t *testing.T) {
	vies := &stubVIES{}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	handler.HandleHealthz(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want ok", resp["status"])
	}
}

func TestExtractPathNumber(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/v1/vat/ES00000000T", "ES00000000T"},
		{"/v1/vat/DE123456789", "DE123456789"},
		{"/v1/vies/ES00000000T", "ES00000000T"},
		{"/v1/vies/DE123456789", "DE123456789"},
		{"/v1/vat/GB123456789", "GB123456789"},
		{"/v1/vat/", ""},
		{"/v1/vies/", ""},
		{"/other/path", ""},
	}

	for _, tt := range tests {
		got := extractPathNumber(tt.path)
		if got != tt.want {
			t.Errorf("extractPathNumber(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestNewRouter_Routes(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	vies := &stubVIES{
		result: &application.VIESResult{Valid: true},
	}

	router := NewRouter(vies, log)
	server := httptest.NewServer(router)
	defer server.Close()

	// Test healthz
	resp, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatalf("healthz request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("healthz status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()

	// Test /v1/vat/ (format only)
	resp, err = http.Get(server.URL + "/v1/vat/GB123456789")
	if err != nil {
		t.Fatalf("vat request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("vat status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()

	// Test /v1/vies/ (with VIES)
	resp, err = http.Get(server.URL + "/v1/vies/DE123456789")
	if err != nil {
		t.Fatalf("vies request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("vies status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

// stubVIES is a test stub for VIESChecker.
type stubVIES struct {
	result *application.VIESResult
	err    error
}

func (s *stubVIES) Check(countryCode, nationalNumber string) (*application.VIESResult, error) {
	return s.result, s.err
}
