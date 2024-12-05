package enum

type CompanyType string

const (
	Corporation        CompanyType = "Corporations"
	NonProfit          CompanyType = "NonProfit"
	Cooperative        CompanyType = "Cooperative"
	SoleProprietorship CompanyType = "Sole Proprietorship"
)

// IsValid validates if the company type is a valid one
func (c CompanyType) IsValid() bool {
	switch c {
	case Corporation, NonProfit, Cooperative, SoleProprietorship:
		return true
	}
	return false
}

// String returns the string representation of the company type
func (c CompanyType) String() string {
	return string(c)
}

// AllCompanyTypes returns all the company types
func AllCompanyTypesString() []string {
	return []string{
		"Corporations",
		"NonProfit",
		"Cooperative",
		"Sole Proprietorship",
	}
}

// CompanyTypeFromString returns the company type from a string
func CompanyTypeFromString(s string) CompanyType {
	switch s {
	case "Corporations":
		return Corporation
	case "NonProfit":
		return NonProfit
	case "Cooperative":
		return Cooperative
	case "Sole Proprietorship":
		return SoleProprietorship
	}
	return ""
}
