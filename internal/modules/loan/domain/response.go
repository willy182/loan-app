package domain

// RequestLoanResponse is the response returned after a loan request.
type RequestLoanResponse struct {
	UserID string      `json:"user_id"`
	Loans  []LoanItem  `json:"loans"`
}

// LoanItem represents a single loan item in the response list.
type LoanItem struct {
	MRP           int64      `json:"mrp"`
	DP            int64      `json:"dp"`
	VehicleYear   int        `json:"vehicle_year"`
	PoliceNumber  string     `json:"police_number"`
	MachineNumber string     `json:"machine_number"`
	Status        LoanStatus `json:"status"`
}

// ApproveLoanSuccessResponse is returned on successful loan approval.
type ApproveLoanSuccessResponse struct {
	UserID       string `json:"user_id"`
	PoliceNumber string `json:"police_number"`
	Message      string `json:"message"`
}

// ApproveLoanErrorResponse is returned when loan approval fails.
type ApproveLoanErrorResponse struct {
	Error           string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
