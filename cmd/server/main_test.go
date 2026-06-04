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

func TestE2E_ValidateVAT_ValidDNI(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/vat-numbers/ES00000000T:validate")
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

	var result infrastructure.ValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestE2E_ValidateVAT_Sanctioned(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/vat-numbers/RU1234567890:validate")
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

func TestE2E_ValidateVAT_NonEU_GB(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/vat-numbers/GB123456789:validate")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestE2E_ValidateVAT_Unparseable(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/v1/vat-numbers/12:validate")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestE2E_ValidateVAT_SpainInvalidControlDigit(t *testing.T) {
	server := createTestServer()
	defer server.Close()

	// 12345678 % 23 = 14 → Z, not A
	resp, err := http.Get(server.URL + "/v1/vat-numbers/ES12345678A:validate")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	// Should be 200 but format_valid should be false
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("response has no data object")
	}
	formatValid, ok := data["format_valid"].(bool)
	if !ok {
		t.Fatal("data has no format_valid field")
	}
	if formatValid {
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
	mux.HandleFunc("GET /v1/vat-numbers/{vatNumber}", handler.HandleValidate)
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
