package enum

// DatabaseType is an enum to represent the possible databases technologies
type DatabaseType string

const (
	Postgres DatabaseType = "postgres"
)

// String returns the string value of the DatabaseType
func (e DatabaseType) String() string {
	switch e {
	case Postgres:
		return "postgres"
	}
	return ""
}

// IsValid checks if the DatabaseType is valid
func (e DatabaseType) IsValid() bool {
	switch e {
	case Postgres:
		return true
	}
	return false
}
