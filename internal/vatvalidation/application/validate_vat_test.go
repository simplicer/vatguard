package application

import (
	"errors"
	"testing"
)

// mockVIESChecker is a mock implementation of VIESChecker for testing.
type mockVIESChecker struct {
	result *VIESResult
	err    error
	called bool
	lastCC string
	lastNN string
}

func (m *mockVIESChecker) Check(countryCode, nationalNumber string) (*VIESResult, error) {
	m.called = true
	m.lastCC = countryCode
	m.lastNN = nationalNumber
	return m.result, m.err
}

func TestValidateVAT_ValidEUMember_VIESRegistered(t *testing.T) {
	vies := &mockVIESChecker{
		result: &VIESResult{
			Valid:       true,
			CompanyName: "TEST COMPANY SL",
			Address:     "CALLE TEST 1, MADRID",
		},
	}

	resp, err := ValidateVAT(ValidateVATRequest{VATNumber: "ESA1234567J"}, vies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.CountryCode != "ES" {
		t.Errorf("CountryCode = %q, want %q", resp.CountryCode, "ES")
	}
	if resp.CountryClassification != "EU_MEMBER" {
		t.Errorf("CountryClassification = %q, want %q", resp.CountryClassification, "EU_MEMBER")
	}
	if !resp.FormatValid {
		t.Error("FormatValid = false, want true")
	}
	if resp.VIESRegistered == nil || !*resp.VIESRegistered {
		t.Error("VIESRegistered = nil or false, want true")
	}
	if resp.VIESCompanyName != "TEST COMPANY SL" {
		t.Errorf("VIESCompanyName = %q, want %q", resp.VIESCompanyName, "TEST COMPANY SL")
	}
	if resp.VIESAddress != "CALLE TEST 1, MADRID" {
		t.Errorf("VIESAddress = %q, want %q", resp.VIESAddress, "CALLE TEST 1, MADRID")
	}
	if !vies.called {
		t.Error("VIES checker was not called")
	}
	if vies.lastCC != "ES" {
		t.Errorf("VIES called with country code %q, want %q", vies.lastCC, "ES")
	}
}

func TestValidateVAT_ValidEUMember_VIESNotRegistered(t *testing.T) {
	vies := &mockVIESChecker{
		result: &VIESResult{Valid: false},
	}

	resp, err := ValidateVAT(ValidateVATRequest{VATNumber: "DE123456789"}, vies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.VIESRegistered == nil || *resp.VIESRegistered {
		t.Error("VIESRegistered should be false")
	}
	if resp.VIESCompanyName != "" {
		t.Errorf("VIESCompanyName should be empty when not registered, got %q", resp.VIESCompanyName)
	}
}

func TestValidateVAT_ValidEUMember_VIESUnavailable(t *testing.T) {
	vies := &mockVIESChecker{
		err: errors.New("connection refused"),
	}

	resp, vErr := ValidateVAT(ValidateVATRequest{VATNumber: "DE123456789"}, vies)
	if vErr == nil {
		t.Fatal("expected VIES unavailable error")
	}
	if vErr.Status != 503 {
		t.Errorf("error status = %d, want 503", vErr.Status)
	}
	if resp == nil {
		t.Fatal("response should not be nil even on VIES error")
	}
	if !resp.FormatValid {
		t.Error("FormatValid should still be true")
	}
}

func TestValidateVAT_InvalidFormat_EUMember(t *testing.T) {
	vies := &mockVIESChecker{}

	resp, err := ValidateVAT(ValidateVATRequest{VATNumber: "DE12345"}, vies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.FormatValid {
		t.Error("FormatValid should be false for invalid format")
	}
	if resp.VIESRegistered != nil {
		t.Error("VIESRegistered should be nil when format is invalid (no VIES call)")
	}
	if vies.called {
		t.Error("VIES should not be called when format is invalid")
	}
}

func TestValidateVAT_NonEURecognized_ValidFormat(t *testing.T) {
	vies := &mockVIESChecker{}

	resp, err := ValidateVAT(ValidateVATRequest{VATNumber: "GB123456789"}, vies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.CountryCode != "GB" {
		t.Errorf("CountryCode = %q, want %q", resp.CountryCode, "GB")
	}
	if resp.CountryClassification != "NON_EU_RECOGNIZED" {
		t.Errorf("CountryClassification = %q, want NON_EU_RECOGNIZED", resp.CountryClassification)
	}
	if !resp.FormatValid {
		t.Error("FormatValid should be true")
	}
	if resp.VIESRegistered != nil {
		t.Error("VIESRegistered should be nil for non-EU countries")
	}
	if vies.called {
		t.Error("VIES should not be called for non-EU countries")
	}
}

func TestValidateVAT_NonEUUnknown(t *testing.T) {
	vies := &mockVIESChecker{}

	resp, err := ValidateVAT(ValidateVATRequest{VATNumber: "XX12345"}, vies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.CountryClassification != "NON_EU_UNKNOWN" {
		t.Errorf("CountryClassification = %q, want NON_EU_UNKNOWN", resp.CountryClassification)
	}
	if resp.FormatValid {
		t.Error("FormatValid should be false for unknown country (no validator)")
	}
}

func TestValidateVAT_SanctionedCountry(t *testing.T) {
	vies := &mockVIESChecker{}

	_, vErr := ValidateVAT(ValidateVATRequest{VATNumber: "RU1234567890"}, vies)
	if vErr == nil {
		t.Fatal("expected sanctioned country error")
	}
	if vErr.Status != 422 {
		t.Errorf("error status = %d, want 422", vErr.Status)
	}
	if vErr.Type != "https://vat-validator.dev/errors/sanctioned-country" {
		t.Errorf("error type = %q, want sanctioned-country", vErr.Type)
	}
	if vies.called {
		t.Error("VIES should not be called for sanctioned countries")
	}
}

func TestValidateVAT_Unparseable_TooShort(t *testing.T) {
	vies := &mockVIESChecker{}

	_, vErr := ValidateVAT(ValidateVATRequest{VATNumber: "A"}, vies)
	if vErr == nil {
		t.Fatal("expected unparseable error")
	}
	if vErr.Status != 422 {
		t.Errorf("error status = %d, want 422", vErr.Status)
	}
	if vErr.Type != "https://vat-validator.dev/errors/unparseable-vat" {
		t.Errorf("error type = %q, want unparseable-vat", vErr.Type)
	}
}

func TestValidateVAT_Unparseable_NonAlphaCode(t *testing.T) {
	vies := &mockVIESChecker{}

	_, vErr := ValidateVAT(ValidateVATRequest{VATNumber: "123456789"}, vies)
	if vErr == nil {
		t.Fatal("expected unparseable error for non-alpha country code")
	}
	if vErr.Status != 422 {
		t.Errorf("error status = %d, want 422", vErr.Status)
	}
}

func TestValidateVAT_Greece_GR_UsesELForVIES(t *testing.T) {
	vies := &mockVIESChecker{
		result: &VIESResult{Valid: true, CompanyName: "GREEK CO"},
	}

	resp, err := ValidateVAT(ValidateVATRequest{VATNumber: "GR123456789"}, vies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.CountryCode != "GR" {
		t.Errorf("CountryCode = %q, want GR", resp.CountryCode)
	}
	if vies.lastCC != "EL" {
		t.Errorf("VIES was called with country code %q, want EL (GR maps to EL for VIES)", vies.lastCC)
	}
}

func TestValidateVAT_ValidatedAtIsSet(t *testing.T) {
	vies := &mockVIESChecker{}

	resp, err := ValidateVAT(ValidateVATRequest{VATNumber: "GB123456789"}, vies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ValidatedAt == "" {
		t.Error("ValidatedAt should be set")
	}
}
