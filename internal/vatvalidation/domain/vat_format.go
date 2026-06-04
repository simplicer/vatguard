package domain

import "strconv"

// VATNumber represents a parsed VAT number with country code and national part.
type VATNumber struct {
	CountryCode    string
	NationalNumber string
}

// ParseVATNumber splits a full VAT number into country code and national number.
// Expects at least 2 characters. The first 2 must be alphabetic (country code).
func ParseVATNumber(input string) (VATNumber, error) {
	if len(input) < 2 {
		return VATNumber{}, errVATTooShort
	}

	code := input[:2]
	national := input[2:]

	if !isAlpha(code[0]) || !isAlpha(code[1]) {
		return VATNumber{}, errCountryCodeNotAlpha
	}

	if len(national) == 0 {
		return VATNumber{}, errVATTooShort
	}

	return VATNumber{
		CountryCode:    toUpper(code),
		NationalNumber: national,
	}, nil
}

// ValidateFormat checks whether a national VAT number matches the expected format
// for the given country code.
func ValidateFormat(countryCode, nationalNumber string) bool {
	code := toUpper(countryCode)

	if IsSanctioned(code) {
		return false
	}

	validator, ok := formatValidators[code]
	if !ok {
		return false
	}
	return validator(nationalNumber)
}

// formatValidators maps country codes to their format validation functions.
var formatValidators = map[string]func(string) bool{
	"AT": validateAustria,
	"BE": validateBelgium,
	"BG": validateBulgaria,
	"HR": validateCroatia,
	"CY": validateCyprus,
	"CZ": validateCzechRepublic,
	"DK": validateDenmark,
	"EE": validateEstonia,
	"FI": validateFinland,
	"FR": validateFrance,
	"DE": validateGermany,
	"EL": validateGreece,
	"GR": validateGreece, // GR uses EL format
	"HU": validateHungary,
	"IE": validateIreland,
	"IT": validateItaly,
	"LV": validateLatvia,
	"LT": validateLithuania,
	"LU": validateLuxembourg,
	"MT": validateMalta,
	"NL": validateNetherlands,
	"PL": validatePoland,
	"PT": validatePortugal,
	"RO": validateRomania,
	"SK": validateSlovakia,
	"SI": validateSlovenia,
	"ES": validateSpain,
	"SE": validateSweden,
	// Non-EU
	"GB": validateUnitedKingdom,
	"NO": validateNorway,
	"CH": validateSwitzerland,
}

// EU validators

func validateAustria(s string) bool {
	// U + 8 digits = 9 chars total
	if len(s) != 9 {
		return false
	}
	return s[0] == 'U' && isDigits(s[1:])
}

func validateBelgium(s string) bool {
	// 0 or 1 + 9 digits = 10 digits total
	if len(s) != 10 {
		return false
	}
	if s[0] != '0' && s[0] != '1' {
		return false
	}
	return isDigits(s)
}

func validateBulgaria(s string) bool {
	// 4-10 digits
	l := len(s)
	return l >= 4 && l <= 10 && isDigits(s)
}

func validateCroatia(s string) bool {
	// 11 digits
	return len(s) == 11 && isDigits(s)
}

func validateCyprus(s string) bool {
	// 8 digits + 1 letter = 9 chars total
	if len(s) != 9 {
		return false
	}
	return isDigits(s[:8]) && isLetter(s[8])
}

func validateCzechRepublic(s string) bool {
	// 8-10 digits
	l := len(s)
	return l >= 8 && l <= 10 && isDigits(s)
}

func validateDenmark(s string) bool {
	// 8 digits
	return len(s) == 8 && isDigits(s)
}

func validateEstonia(s string) bool {
	// 9 digits
	return len(s) == 9 && isDigits(s)
}

func validateFinland(s string) bool {
	// 8 digits
	return len(s) == 8 && isDigits(s)
}

func validateFrance(s string) bool {
	// 2 alphanumeric chars + 9 digits = 11 chars total
	if len(s) != 11 {
		return false
	}
	return isAlphanumeric(s[:2]) && isDigits(s[2:])
}

func validateGermany(s string) bool {
	// 9 digits
	return len(s) == 9 && isDigits(s)
}

func validateGreece(s string) bool {
	// 9 digits
	return len(s) == 9 && isDigits(s)
}

func validateHungary(s string) bool {
	// 8 digits
	return len(s) == 8 && isDigits(s)
}

func validateIreland(s string) bool {
	// Old: 7 digits + 1 letter (A-W) = 8 chars
	// New: 7 digits + 2 letters = 9 chars
	l := len(s)
	if l == 8 {
		return isDigits(s[:7]) && isLetter(s[7])
	}
	if l == 9 {
		return isDigits(s[:7]) && isLetter(s[7]) && isLetter(s[8])
	}
	return false
}

func validateItaly(s string) bool {
	// 11 digits
	return len(s) == 11 && isDigits(s)
}

func validateLatvia(s string) bool {
	// 11 digits
	return len(s) == 11 && isDigits(s)
}

func validateLithuania(s string) bool {
	// 9 or 12 digits
	l := len(s)
	return (l == 9 || l == 12) && isDigits(s)
}

func validateLuxembourg(s string) bool {
	// 8 digits
	return len(s) == 8 && isDigits(s)
}

func validateMalta(s string) bool {
	// 6 digits (MT prefix is already stripped)
	return len(s) == 6 && isDigits(s)
}

func validateNetherlands(s string) bool {
	// 9 digits + B + 2 digits = 12 chars total
	if len(s) != 12 {
		return false
	}
	return isDigits(s[:9]) && s[9] == 'B' && isDigits(s[10:])
}

func validatePoland(s string) bool {
	// 10 digits
	return len(s) == 10 && isDigits(s)
}

func validatePortugal(s string) bool {
	// 9 digits
	return len(s) == 9 && isDigits(s)
}

func validateRomania(s string) bool {
	// 2-10 digits
	l := len(s)
	return l >= 2 && l <= 10 && isDigits(s)
}

func validateSlovakia(s string) bool {
	// 10 digits, must be divisible by 11
	if len(s) != 10 {
		return false
	}
	if !isDigits(s) {
		return false
	}
	num, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return num%11 == 0
}

func validateSlovenia(s string) bool {
	// 8 digits
	return len(s) == 8 && isDigits(s)
}

func validateSpain(s string) bool {
	// Letter + 7 alphanumeric chars + letter/digit = 9 chars total
	if len(s) != 9 {
		return false
	}
	if !isLetter(s[0]) {
		return false
	}
	for i := 1; i <= 7; i++ {
		if !isAlphanumericByte(s[i]) {
			return false
		}
	}
	return isLetter(s[8]) || isDigit(s[8])
}

func validateSweden(s string) bool {
	// 12 digits, last 2 must be 01-12 or 99
	if len(s) != 12 {
		return false
	}
	if !isDigits(s) {
		return false
	}
	suffix := s[10:]
	val, err := strconv.Atoi(suffix)
	if err != nil {
		return false
	}
	return (val >= 1 && val <= 12) || val == 99
}

// Non-EU validators

func validateUnitedKingdom(s string) bool {
	// 9 digits (standard), 12 digits (branch), GD + 3 digits, HA + 3 digits
	l := len(s)
	if l == 9 && isDigits(s) {
		return true
	}
	if l == 12 && isDigits(s) {
		return true
	}
	if l == 5 && (s[:2] == "GD" || s[:2] == "HA") && isDigits(s[2:]) {
		return true
	}
	return false
}

func validateNorway(s string) bool {
	// 9 digits
	return len(s) == 9 && isDigits(s)
}

func validateSwitzerland(s string) bool {
	// CHE + 9 digits
	if len(s) != 12 {
		return false
	}
	return s[:3] == "CHE" && isDigits(s[3:])
}

// Error types

type parseError struct {
	msg string
}

func (e *parseError) Error() string {
	return e.msg
}

var (
	errVATTooShort         = &parseError{msg: "VAT number is too short"}
	errCountryCodeNotAlpha = &parseError{msg: "country code must be alphabetic"}
)

// Helper functions

func isAlpha(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isDigit(s[i]) {
			return false
		}
	}
	return true
}

func isAlphanumericByte(c byte) bool {
	return isDigit(c) || isLetter(c)
}

func isAlphanumeric(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isAlphanumericByte(s[i]) {
			return false
		}
	}
	return true
}
