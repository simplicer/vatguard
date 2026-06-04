package application

import (
	"time"

	"codeberg.org/simplicer/vatguard/internal/vatvalidation/domain"
)

// VIESResult contains the result of a VIES lookup.
type VIESResult struct {
	Valid       bool
	CompanyName string
	Address     string
}

// VIESChecker is the port for VIES verification.
type VIESChecker interface {
	Check(countryCode, nationalNumber string) (*VIESResult, error)
}

// ValidateVATRequest is the input for the validate VAT use case.
type ValidateVATRequest struct {
	VATNumber   string
	IncludeVIES bool
}

// ValidateVATResponse is the output of the validate VAT use case.
type ValidateVATResponse struct {
	VATNumber             string                       `json:"vat_number"`
	CountryCode           string                       `json:"country_code"`
	CountryClassification domain.CountryClassification `json:"country_classification"`
	FormatValid           bool                         `json:"format_valid"`
	VIESRegistered        *bool                        `json:"vies_registered"`
	VIESCompanyName       string                       `json:"vies_company_name,omitempty"`
	VIESAddress           string                       `json:"vies_address,omitempty"`
	ValidatedAt           string                       `json:"validated_at"`
}

// ValidationError represents a business-rule validation error.
type ValidationError struct {
	Type   string
	Title  string
	Detail string
	Status int
}

func (e *ValidationError) Error() string {
	return e.Title
}

// ValidateVAT executes the VAT validation use case.
func ValidateVAT(req ValidateVATRequest, vies VIESChecker) (*ValidateVATResponse, *ValidationError) {
	parsed, err := domain.ParseVATNumber(req.VATNumber)
	if err != nil {
		return nil, &ValidationError{
			Type:   "https://vat-validator.dev/errors/unparseable-vat",
			Title:  "Unparseable VAT number",
			Detail: "The provided VAT number could not be parsed. Expected format: two-letter country code followed by the national VAT number.",
			Status: 422,
		}
	}

	classification := domain.ClassifyCountry(parsed.CountryCode)

	if classification == domain.Sanctioned {
		return nil, &ValidationError{
			Type:   "https://vat-validator.dev/errors/sanctioned-country",
			Title:  "Sanctioned country",
			Detail: "VAT validation is not available for countries under EU sanctions: " + parsed.CountryCode,
			Status: 422,
		}
	}

	formatValid := domain.ValidateFormat(parsed.CountryCode, parsed.NationalNumber)

	resp := &ValidateVATResponse{
		VATNumber:             req.VATNumber,
		CountryCode:           parsed.CountryCode,
		CountryClassification: classification,
		FormatValid:           formatValid,
		ValidatedAt:           time.Now().UTC().Format(time.RFC3339),
	}

	// VIES lookup only when explicitly requested, for EU members with valid format
	if req.IncludeVIES && classification == domain.EUMember && formatValid {
		viesCode := domain.GetVIESCountryCode(parsed.CountryCode)
		if viesCode != "" {
			result, err := vies.Check(viesCode, parsed.NationalNumber)
			if err != nil {
				return resp, &ValidationError{
					Type:   "https://vat-validator.dev/errors/vies-unavailable",
					Title:  "VIES service unavailable",
					Detail: "Could not reach the VIES SOAP service to verify the VAT number",
					Status: 503,
				}
			}
			viesValid := result.Valid
			resp.VIESRegistered = &viesValid
			if result.Valid {
				resp.VIESCompanyName = result.CompanyName
				resp.VIESAddress = result.Address
			}
		}
	}

	return resp, nil
}
