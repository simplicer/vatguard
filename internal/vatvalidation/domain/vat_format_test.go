package domain

import "testing"

func TestParseVATNumber_Valid(t *testing.T) {
	tests := []struct {
		input        string
		wantCountry  string
		wantNational string
	}{
		{"ESA12345674", "ES", "A12345674"},
		{"DE123456789", "DE", "123456789"},
		{"FRAB123456789", "FR", "AB123456789"},
		{"NL123456789B12", "NL", "123456789B12"},
		{"GB123456789", "GB", "123456789"},
	}

	for _, tt := range tests {
		parsed, err := ParseVATNumber(tt.input)
		if err != nil {
			t.Errorf("ParseVATNumber(%q) returned error: %v", tt.input, err)
			continue
		}
		if parsed.CountryCode != tt.wantCountry {
			t.Errorf("ParseVATNumber(%q).CountryCode = %q, want %q", tt.input, parsed.CountryCode, tt.wantCountry)
		}
		if parsed.NationalNumber != tt.wantNational {
			t.Errorf("ParseVATNumber(%q).NationalNumber = %q, want %q", tt.input, parsed.NationalNumber, tt.wantNational)
		}
	}
}

func TestParseVATNumber_TooShort(t *testing.T) {
	_, err := ParseVATNumber("E")
	if err == nil {
		t.Error("ParseVATNumber(\"E\") should return error for too short input")
	}
}

func TestParseVATNumber_Empty(t *testing.T) {
	_, err := ParseVATNumber("")
	if err == nil {
		t.Error("ParseVATNumber(\"\") should return error for empty input")
	}
}

func TestParseVATNumber_CountryCodeNotAlpha(t *testing.T) {
	_, err := ParseVATNumber("123456789")
	if err == nil {
		t.Error("ParseVATNumber(\"123456789\") should return error for non-alpha country code")
	}
}

func TestValidateFormat_Spain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		// NIF (DNI): 8 digits + control letter (mod 23)
		{"valid DNI all zeros", "00000000T", true},       // 0 % 23 = 0 → T
		{"valid DNI 12345678", "12345678Z", true},        // 12345678 % 23 = 14 → Z
		{"invalid DNI wrong letter", "12345678A", false}, // should be Z, not A
		{"invalid DNI digit at end", "000000009", false}, // must end with letter

		// NIE: X/Y/Z + 7 digits + control letter
		{"valid NIE X", "X0000000T", true},               // X→0, 00000000 % 23 = 0 → T
		{"valid NIE Y", "Y0000000Z", true},               // Y→1, 10000000 % 23 = 14 → Z
		{"valid NIE Z", "Z0000000M", true},               // Z→2, 20000000 % 23 = 5 → M
		{"valid NIE X2345678", "X2345678T", true},        // X→0, 02345678 % 23 = 0 → T
		{"invalid NIE wrong letter", "X0000000A", false}, // should be T, not A
		{"invalid NIE prefix W", "W0000000T", false},     // W is not a valid NIE prefix

		// NIF jurídica: type letter + 7 digits + control
		{"valid B (SL)", "B12345674", true},             // B=digit control, 1234567 → 4
		{"valid A (SA)", "A12345674", true},             // A=digit control, 1234567 → 4
		{"valid G (Asoc)", "G00000000", true},           // G=digit control, 0000000 → 0
		{"valid S (Admin)", "S0000000J", true},          // S=letter control, 0000000→0→J
		{"valid P (Local)", "P0000000J", true},          // P=letter control, 0000000→0→J
		{"valid N (extranjera)", "N0000000J", true},     // N=letter control, 0000000→0→J
		{"invalid B wrong digit", "B12345675", false},   // should be 4, not 5
		{"invalid S wrong letter", "S0000000A", false},  // should be J, not A
		{"invalid entity letter I", "I12345678", false}, // I is not a valid entity type
		{"invalid entity letter T", "T12345678", false}, // T is not a valid entity type

		// Length/format checks
		{"too short", "1234567", false},   // only 7 chars
		{"too long", "1234567890", false}, // 10 chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateFormat("ES", tt.input)
			if result != tt.valid {
				t.Errorf("ValidateFormat(\"ES\", %q) = %v, want %v", tt.input, result, tt.valid)
			}
		})
	}
}

func TestValidateFormat_Germany(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"123456789", true},
		{"12345678", false},   // too short
		{"1234567890", false}, // too long
		{"A23456789", false},  // contains letter
	}

	for _, tt := range tests {
		result := ValidateFormat("DE", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"DE\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_France(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"AB123456789", true},
		{"00123456789", true},
		{"123456789", false},    // too short - missing leading chars
		{"AB1234567890", false}, // too long
	}

	for _, tt := range tests {
		result := ValidateFormat("FR", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"FR\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Netherlands(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"123456789B12", true},
		{"123456789B01", true},
		{"123456789C12", false}, // C instead of B
		{"12345678B12", false},  // only 8 digits
		{"123456789B1", false},  // only 1 digit after B
	}

	for _, tt := range tests {
		result := ValidateFormat("NL", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"NL\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Italy(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678901", true},
		{"1234567890", false},   // 10 digits
		{"123456789012", false}, // 12 digits
		{"A2345678901", false},  // contains letter
	}

	for _, tt := range tests {
		result := ValidateFormat("IT", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"IT\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Austria(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"U12345678", true},
		{"U1234567", false},   // too short
		{"U123456789", false}, // too long
		{"A12345678", false},  // doesn't start with U
		{"12345678", false},   // doesn't start with U
	}

	for _, tt := range tests {
		result := ValidateFormat("AT", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"AT\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Belgium(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"0123456789", true},
		{"1123456789", true},
		{"2123456789", false},  // starts with 2
		{"012345678", false},   // only 9 digits
		{"01234567890", false}, // 11 digits
	}

	for _, tt := range tests {
		result := ValidateFormat("BE", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"BE\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Bulgaria(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"1234", true},
		{"1234567890", true},
		{"123", false},         // too short
		{"12345678901", false}, // too long
	}

	for _, tt := range tests {
		result := ValidateFormat("BG", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"BG\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Croatia(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678901", true},
		{"1234567890", false},   // 10 digits
		{"123456789012", false}, // 12 digits
	}

	for _, tt := range tests {
		result := ValidateFormat("HR", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"HR\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Cyprus(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678A", true},
		{"123456789", false},  // ends with digit not letter
		{"1234567A", false},   // too short
		{"12345678AB", false}, // too long
	}

	for _, tt := range tests {
		result := ValidateFormat("CY", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"CY\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_CzechRepublic(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678", true},
		{"123456789", true},
		{"1234567890", true},
		{"1234567", false},     // 7 digits
		{"12345678901", false}, // 11 digits
	}

	for _, tt := range tests {
		result := ValidateFormat("CZ", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"CZ\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Denmark(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678", true},
		{"1234567", false},   // too short
		{"123456789", false}, // too long
	}

	for _, tt := range tests {
		result := ValidateFormat("DK", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"DK\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Estonia(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"123456789", true},
		{"12345678", false},   // too short
		{"1234567890", false}, // too long
	}

	for _, tt := range tests {
		result := ValidateFormat("EE", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"EE\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Finland(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678", true},
		{"1234567", false},   // too short
		{"123456789", false}, // too long
	}

	for _, tt := range tests {
		result := ValidateFormat("FI", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"FI\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Greece(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"123456789", true},
		{"12345678", false},   // too short
		{"1234567890", false}, // too long
	}

	for _, tt := range tests {
		result := ValidateFormat("EL", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"EL\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_GRUsesELFormat(t *testing.T) {
	result := ValidateFormat("GR", "123456789")
	if !result {
		t.Error("ValidateFormat(\"GR\", \"123456789\") should use EL format and return true")
	}
}

func TestValidateFormat_Hungary(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678", true},
		{"1234567", false},   // too short
		{"123456789", false}, // too long
	}

	for _, tt := range tests {
		result := ValidateFormat("HU", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"HU\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Ireland(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"1234567A", true},   // old format: 7 digits + 1 letter
		{"1234567AB", true},  // new format: 7 digits + 2 letters
		{"12345678", false},  // 8 digits, no letter
		{"123456", false},    // too short
		{"12345678A", false}, // 8 digits + letter
	}

	for _, tt := range tests {
		result := ValidateFormat("IE", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"IE\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Latvia(t *testing.T) {
	result := ValidateFormat("LV", "12345678901")
	if !result {
		t.Error("ValidateFormat(\"LV\", \"12345678901\") should be true (11 digits)")
	}
}

func TestValidateFormat_Lithuania(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"123456789", true},    // 9 digits
		{"123456789012", true}, // 12 digits
		{"12345678", false},    // 8 digits
		{"1234567890", false},  // 10 digits
		{"12345678901", false}, // 11 digits
	}

	for _, tt := range tests {
		result := ValidateFormat("LT", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"LT\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Luxembourg(t *testing.T) {
	result := ValidateFormat("LU", "12345678")
	if !result {
		t.Error("ValidateFormat(\"LU\", \"12345678\") should be true (8 digits)")
	}
}

func TestValidateFormat_Malta(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"123456", true},   // 6 digits (MT prefix stripped)
		{"12345", false},   // 5 digits
		{"1234567", false}, // 7 digits
	}

	for _, tt := range tests {
		result := ValidateFormat("MT", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"MT\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Poland(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"1234567890", true},
		{"123456789", false},   // 9 digits
		{"12345678901", false}, // 11 digits
	}

	for _, tt := range tests {
		result := ValidateFormat("PL", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"PL\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Portugal(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"123456789", true},
		{"12345678", false},   // too short
		{"1234567890", false}, // too long
	}

	for _, tt := range tests {
		result := ValidateFormat("PT", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"PT\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Romania(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12", true},           // minimum 2 digits
		{"1234567890", true},   // maximum 10 digits
		{"1", false},           // too short
		{"12345678901", false}, // too long
		{"A234567890", false},  // contains letter
	}

	for _, tt := range tests {
		result := ValidateFormat("RO", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"RO\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Slovakia(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"2022744009", true},   // 2022744009 / 11 = 183885819
		{"1234567890", false},  // not divisible by 11
		{"0000000011", true},   // 11, divisible by 11
		{"123456789", false},   // 9 digits
		{"12345678901", false}, // 11 digits
	}

	for _, tt := range tests {
		result := ValidateFormat("SK", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"SK\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Slovenia(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678", true},
		{"1234567", false},   // too short
		{"123456789", false}, // too long
	}

	for _, tt := range tests {
		result := ValidateFormat("SI", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"SI\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_Sweden(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"123456789901", true},   // ends with 01
		{"123456789912", true},   // ends with 12
		{"123456789999", true},   // ends with 99
		{"123456789900", false},  // ends with 00 (invalid)
		{"123456789913", false},  // ends with 13 (invalid)
		{"1234567899", false},    // 10 digits
		{"1234567899012", false}, // 13 digits
	}

	for _, tt := range tests {
		result := ValidateFormat("SE", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"SE\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_UnknownCountry(t *testing.T) {
	result := ValidateFormat("XX", "12345")
	if result {
		t.Error("ValidateFormat(\"XX\", \"12345\") should return false for unknown country")
	}
}

func TestValidateFormat_SanctionedCountry(t *testing.T) {
	result := ValidateFormat("RU", "1234567890")
	if result {
		t.Error("ValidateFormat(\"RU\", \"1234567890\") should return false for sanctioned country")
	}
}

func TestValidateFormat_NonEU_UnitedKingdom(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"123456789", true},    // 9 digits (standard)
		{"123456789012", true}, // 12 digits (branch traders)
		{"GD123", true},        // government department
		{"HA123", true},        // health authority
		{"12345678", false},    // 8 digits
		{"12345678901", false}, // 11 digits
	}

	for _, tt := range tests {
		result := ValidateFormat("GB", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"GB\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestValidateFormat_NonEU_Norway(t *testing.T) {
	result := ValidateFormat("NO", "123456789")
	if !result {
		t.Error("ValidateFormat(\"NO\", \"123456789\") should be true (9 digits)")
	}
}

func TestValidateFormat_NonEU_Switzerland(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"CHE123456789", true},
		{"CH123456789", false}, // doesn't start with CHE
		{"CHE12345678", false}, // only 8 digits after CHE
	}

	for _, tt := range tests {
		result := ValidateFormat("CH", tt.input)
		if result != tt.valid {
			t.Errorf("ValidateFormat(\"CH\", %q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}
