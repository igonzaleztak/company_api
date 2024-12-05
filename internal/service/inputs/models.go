package inputs

// CreateCompany represents the input for creating a company
type CreateCompanyInput struct {
	Name            string `json:"name" validate:"required"`
	Description     string `json:"description,omitempty"`
	AmountEmployees *int   `json:"amount_employees" validate:"required"`
	Registered      *bool  `json:"registered" validate:"required"`
	Type            string `json:"type" validate:"required,oneof=Corporations NonProfit Cooperative sole_proprietorship"`
}

// UpdateCompany represents the input for updating a company
type UpdateCompany CreateCompanyInput
