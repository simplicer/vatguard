package infrastructure

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/antonio-villamarin/vat-validator/internal/vatvalidation/application"
)

func TestHandleValidate_Success_EU(t *testing.T) {
	vies := &stubVIES{
		result: &application.VIESResult{
			Valid:       true,
			CompanyName: "TEST SL",
			Address:     "ADDR",
		},
	}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vat-numbers/ESA12345674:validate", nil)
	w := httptest.NewRecorder()

	handler.HandleValidate(w, req)

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
}

func TestHandleValidate_SanctionedCountry(t *testing.T) {
	vies := &stubVIES{}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vat-numbers/RU1234567890:validate", nil)
	w := httptest.NewRecorder()

	handler.HandleValidate(w, req)

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

func TestHandleValidate_Unparseable(t *testing.T) {
	vies := &stubVIES{}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vat-numbers/1:validate", nil)
	w := httptest.NewRecorder()

	handler.HandleValidate(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

func TestHandleValidate_VIESUnavailable(t *testing.T) {
	vies := &stubVIES{err: errors.New("unavailable")}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vat-numbers/DE123456789:validate", nil)
	w := httptest.NewRecorder()

	handler.HandleValidate(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleValidate_NonEU(t *testing.T) {
	vies := &stubVIES{}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	req := httptest.NewRequest(http.MethodGet, "/v1/vat-numbers/GB123456789:validate", nil)
	w := httptest.NewRecorder()

	handler.HandleValidate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleHealthz(t *testing.T) {
	vies := &stubVIES{}
	handler := NewHTTPHandler(vies, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

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

func TestExtractVATNumber(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/v1/vat-numbers/ESA12345674:validate", "ESA12345674"},
		{"/v1/vat-numbers/DE123456789:validate", "DE123456789"},
		{"/v1/vat-numbers/GB123456789", "GB123456789"},
		{"/v1/vat-numbers/:validate", ""},
		{"/v1/vat-numbers/", ""},
		{"/other/path", ""},
	}

	for _, tt := range tests {
		got := extractVATNumber(tt.path)
		if got != tt.want {
			t.Errorf("extractVATNumber(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestNewRouter_Routes(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
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
}

// stubVIES is a test stub for VIESChecker.
type stubVIES struct {
	result *application.VIESResult
	err    error
}

func (s *stubVIES) Check(countryCode, nationalNumber string) (*application.VIESResult, error) {
	return s.result, s.err
}
