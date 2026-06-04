package infrastructure

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVIESClient_Check_ValidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <checkVatResponse xmlns="urn:ec.europa.eu:taxud:vies:services:checkVat">
      <countryCode>ES</countryCode>
      <vatNumber>A1234567J</vatNumber>
      <valid>true</valid>
      <name>TEST COMPANY SL</name>
      <address>CALLE TEST 1, MADRID</address>
    </checkVatResponse>
  </soapenv:Body>
</soapenv:Envelope>`))
	}))
	defer server.Close()

	client := NewVIESClientWithURL(server.URL)
	result, err := client.Check("ES", "A1234567J")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("result.Valid = false, want true")
	}
	if result.CompanyName != "TEST COMPANY SL" {
		t.Errorf("CompanyName = %q, want %q", result.CompanyName, "TEST COMPANY SL")
	}
	if result.Address != "CALLE TEST 1, MADRID" {
		t.Errorf("Address = %q, want %q", result.Address, "CALLE TEST 1, MADRID")
	}
}

func TestVIESClient_Check_InvalidVAT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <checkVatResponse xmlns="urn:ec.europa.eu:taxud:vies:services:checkVat">
      <countryCode>DE</countryCode>
      <vatNumber>123456789</vatNumber>
      <valid>false</valid>
      <name>---</name>
      <address>---</address>
    </checkVatResponse>
  </soapenv:Body>
</soapenv:Envelope>`))
	}))
	defer server.Close()

	client := NewVIESClientWithURL(server.URL)
	result, err := client.Check("DE", "123456789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("result.Valid = true, want false")
	}
}

func TestVIESClient_Check_SoapFault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <soapenv:Fault>
      <faultstring>INVALID_INPUT</faultstring>
    </soapenv:Fault>
  </soapenv:Body>
</soapenv:Envelope>`))
	}))
	defer server.Close()

	client := NewVIESClientWithURL(server.URL)
	_, err := client.Check("ES", "INVALID")
	if err == nil {
		t.Fatal("expected error for SOAP fault")
	}
}

func TestVIESClient_Check_ServerUnreachable(t *testing.T) {
	client := NewVIESClientWithURL("http://localhost:1")
	_, err := client.Check("ES", "A1234567J")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestVIESClient_Check_Non200Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewVIESClientWithURL(server.URL)
	_, err := client.Check("ES", "A1234567J")
	if err == nil {
		t.Fatal("expected error for HTTP 503")
	}
}

func TestVIESClient_Check_ContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "text/xml; charset=utf-8" {
			t.Errorf("Content-Type = %q, want %q", ct, "text/xml; charset=utf-8")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <checkVatResponse xmlns="urn:ec.europa.eu:taxud:vies:services:checkVat">
      <countryCode>ES</countryCode>
      <vatNumber>A1234567J</vatNumber>
      <valid>true</valid>
      <name>TEST</name>
      <address>ADDR</address>
    </checkVatResponse>
  </soapenv:Body>
</soapenv:Envelope>`))
	}))
	defer server.Close()

	client := NewVIESClientWithURL(server.URL)
	_, _ = client.Check("ES", "A1234567J")
}

func TestParseVIESResponse_InvalidXML(t *testing.T) {
	_, err := parseVIESResponse([]byte("not xml"))
	if err == nil {
		t.Fatal("expected error for invalid XML")
	}
}

func TestXmlEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ES", "ES"},
		{"A&B", "A&amp;B"},
		{"A<B", "A&lt;B"},
		{"A>B", "A&gt;B"},
		{`A"B`, "A&quot;B"},
		{"A'B", "A&apos;B"},
	}

	for _, tt := range tests {
		result := xmlEscape(tt.input)
		if result != tt.expected {
			t.Errorf("xmlEscape(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
