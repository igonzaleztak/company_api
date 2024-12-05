package options

// DatabaseOptions represents the options for the database
type DatabaseOptions struct {
	ConnString string
}

// WithConnectionString is a functional option to set the connection string
func WithConnectionString(connStr string) func(*DatabaseOptions) {
	return func(o *DatabaseOptions) {
		o.ConnString = connStr
	}
}
