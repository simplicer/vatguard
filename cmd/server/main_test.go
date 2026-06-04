package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antonio-villamarin/vat-validator/internal/vatvalidation/application"
	"github.com/antonio-villamarin/vat-validator/internal/vatvalidation/infrastructure"
)

func TestE2E_Healthz(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatalf("healthz request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want ok", body["status"])
	}
}

func TestE2E_VAT_FormatOnly(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/vat/ES00000000T")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}

	data := result["data"].(map[string]interface{})
	if data["format_valid"] != true {
		t.Error("format_valid should be true")
	}
	if data["vies_registered"] != nil {
		t.Error("vies_registered should be nil on /vat endpoint (no VIES call)")
	}
}

func TestE2E_VIES_WithVIESLookup(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/vies/ES00000000T")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}

	data := result["data"].(map[string]interface{})
	if data["vies_registered"] == nil {
		t.Error("vies_registered should NOT be nil on /vies endpoint")
	}
}

func TestE2E_VAT_Sanctioned(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/vat/RU1234567890")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
}

func TestE2E_VAT_NonEU_GB(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/vat/GB123456789")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestE2E_VAT_Unparseable(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/vat/12")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestE2E_VAT_SpainInvalidControlDigit(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/vat/ES12345678A")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}

	data := result["data"].(map[string]interface{})
	if data["format_valid"] == true {
		t.Error("format_valid should be false for invalid Spanish NIF control digit")
	}
}

func createTestServer() *httptest.Server {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	vies := &stubVIES{
		result: &application.VIESResult{
			Valid:       true,
			CompanyName: "TEST COMPANY",
			Address:     "TEST ADDRESS",
		},
	}
	handler := infrastructure.NewHTTPHandler(vies, logger)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/vat/{number}", handler.HandleVAT)
	mux.HandleFunc("GET /v1/vies/{number}", handler.HandleVIES)
	mux.HandleFunc("GET /healthz", handler.HandleHealthz)
	return httptest.NewServer(mux)
}

type stubVIES struct {
	result *application.VIESResult
	err    error
}

func (s *stubVIES) Check(countryCode, nationalNumber string) (*application.VIESResult, error) {
	return s.result, s.err
}
