package domain

// RequestLoanRequest is the input for requesting a new loan.
type RequestLoanRequest struct {
	UserID        string `json:"user_id"`
	MRP           int64  `json:"mrp"`
	DP            int64  `json:"dp"`
	VehicleYear   int    `json:"vehicle_year"`
	PoliceNumber  string `json:"police_number"`
	MachineNumber string `json:"machine_number"`
}

// ApproveLoanRequest is the input for approving an existing loan.
type ApproveLoanRequest struct {
	UserID       string `json:"user_id"`
	PoliceNumber string `json:"police_number"`
}
