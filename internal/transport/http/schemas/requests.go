package schemas

// RegisterRequest is the request schema for registering a new user
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginRequest is the request schema for logging in a user
type LoginRequest RegisterRequest

// CreateCompanyRequest is the request schema for creating a company
type CreateCompanyRequest struct {
	Name            string `json:"name" validate:"required"`
	Description     string `json:"description,omitempty"`
	AmountEmployees *int   `json:"amount_employees" validate:"required"`
	Registered      *bool  `json:"registered" validate:"required"`
	Type            string `json:"type" validate:"required,customOneOf"`
}

// UpdateCompanyRequest is the request schema for updating a company
type UpdateCompanyRequest CreateCompanyRequest
