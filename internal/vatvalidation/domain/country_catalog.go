package domain

type CountryClassification string

const (
	EUMember        CountryClassification = "EU_MEMBER"
	NonEURecognized CountryClassification = "NON_EU_RECOGNIZED"
	NonEUUnknown    CountryClassification = "NON_EU_UNKNOWN"
	Sanctioned      CountryClassification = "SANCTIONED"
)

var euMembers = map[string]bool{
	"AT": true, "BE": true, "BG": true, "HR": true, "CY": true,
	"CZ": true, "DK": true, "EE": true, "FI": true, "FR": true,
	"DE": true, "EL": true, "GR": true, "HU": true, "IE": true,
	"IT": true, "LV": true, "LT": true, "LU": true, "MT": true,
	"NL": true, "PL": true, "PT": true, "RO": true, "SK": true,
	"SI": true, "ES": true, "SE": true,
}

var sanctionedCountries = map[string]bool{
	"RU": true, "BY": true,
}

var nonEURecognized = map[string]bool{
	"GB": true, "NO": true, "CH": true,
	"AU": true, "JP": true, "TR": true,
	"CA": true, "CN": true, "IN": true,
	"KR": true, "MX": true, "ZA": true,
	"BR": true, "AR": true,
}

var viesCountryCodes = map[string]string{
	"AT": "AT", "BE": "BE", "BG": "BG", "HR": "HR", "CY": "CY",
	"CZ": "CZ", "DK": "DK", "EE": "EE", "FI": "FI", "FR": "FR",
	"DE": "DE", "EL": "EL", "GR": "EL", "HU": "HU", "IE": "IE",
	"IT": "IT", "LV": "LV", "LT": "LT", "LU": "LU", "MT": "MT",
	"NL": "NL", "PL": "PL", "PT": "PT", "RO": "RO", "SK": "SK",
	"SI": "SI", "ES": "ES", "SE": "SE",
}

func ClassifyCountry(code string) CountryClassification {
	upper := toUpper(code)

	if euMembers[upper] {
		return EUMember
	}
	if sanctionedCountries[upper] {
		return Sanctioned
	}
	if nonEURecognized[upper] {
		return NonEURecognized
	}
	return NonEUUnknown
}

func IsEUMember(code string) bool {
	return euMembers[toUpper(code)]
}

func IsSanctioned(code string) bool {
	return sanctionedCountries[toUpper(code)]
}

func GetVIESCountryCode(code string) string {
	return viesCountryCodes[toUpper(code)]
}

func toUpper(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 32
		}
	}
	return string(b)
}
