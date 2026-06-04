package domain

import "testing"

func TestClassifyCountry_EUMembers(t *testing.T) {
	euCodes := []string{
		"AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR",
		"DE", "EL", "HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL",
		"PL", "PT", "RO", "SK", "SI", "ES", "SE",
	}

	for _, code := range euCodes {
		classification := ClassifyCountry(code)
		if classification != EUMember {
			t.Errorf("ClassifyCountry(%q) = %q, want %q", code, classification, EUMember)
		}
	}
}

func TestClassifyCountry_GRIsEUMember(t *testing.T) {
	classification := ClassifyCountry("GR")
	if classification != EUMember {
		t.Errorf("ClassifyCountry(\"GR\") = %q, want %q (GR maps to EL for VIES but is EU member)", classification, EUMember)
	}
}

func TestClassifyCountry_Sanctioned(t *testing.T) {
	sanctionedCodes := []string{"RU", "BY"}

	for _, code := range sanctionedCodes {
		classification := ClassifyCountry(code)
		if classification != Sanctioned {
			t.Errorf("ClassifyCountry(%q) = %q, want %q", code, classification, Sanctioned)
		}
	}
}

func TestClassifyCountry_NonEURecognized(t *testing.T) {
	nonEUCodes := []string{"GB", "NO", "CH", "AU", "JP", "TR", "CA", "CN", "IN", "KR", "MX", "ZA", "BR", "AR"}

	for _, code := range nonEUCodes {
		classification := ClassifyCountry(code)
		if classification != NonEURecognized {
			t.Errorf("ClassifyCountry(%q) = %q, want %q", code, classification, NonEURecognized)
		}
	}
}

func TestClassifyCountry_Unknown(t *testing.T) {
	unknownCodes := []string{"XX", "ZZ", "AB", "CD", "EF", "GH"}

	for _, code := range unknownCodes {
		classification := ClassifyCountry(code)
		if classification != NonEUUnknown {
			t.Errorf("ClassifyCountry(%q) = %q, want %q", code, classification, NonEUUnknown)
		}
	}
}

func TestClassifyCountry_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected CountryClassification
	}{
		{"es", EUMember},
		{"Es", EUMember},
		{"eS", EUMember},
		{"ru", Sanctioned},
		{"Ru", Sanctioned},
		{"gb", NonEURecognized},
		{"Gb", NonEURecognized},
		{"xx", NonEUUnknown},
	}

	for _, tt := range tests {
		classification := ClassifyCountry(tt.input)
		if classification != tt.expected {
			t.Errorf("ClassifyCountry(%q) = %q, want %q", tt.input, classification, tt.expected)
		}
	}
}

func TestIsEUMember(t *testing.T) {
	if !IsEUMember("ES") {
		t.Error("IsEUMember(\"ES\") = false, want true")
	}
	if IsEUMember("GB") {
		t.Error("IsEUMember(\"GB\") = true, want false")
	}
	if IsEUMember("RU") {
		t.Error("IsEUMember(\"RU\") = true, want false")
	}
}

func TestIsEUMember_GR(t *testing.T) {
	if !IsEUMember("GR") {
		t.Error("IsEUMember(\"GR\") = false, want true (GR is Greece, EU member)")
	}
}

func TestGetVIESCountryCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ES", "ES"},
		{"DE", "DE"},
		{"GR", "EL"},
		{"EL", "EL"},
		{"GB", ""},
		{"RU", ""},
	}

	for _, tt := range tests {
		result := GetVIESCountryCode(tt.input)
		if result != tt.expected {
			t.Errorf("GetVIESCountryCode(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
