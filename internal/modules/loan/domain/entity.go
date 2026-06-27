package domain

import (
	"errors"
	"time"
)

// Sentinel errors for expected domain conditions.
var (
	ErrLoanNotFound = errors.New("loan_not_found")
)

// LoanStatus represents the current state of a loan application.
type LoanStatus string

const (
	LoanStatusSubmitted LoanStatus = "submitted"
	LoanStatusApproved  LoanStatus = "approved"
	LoanStatusRejected  LoanStatus = "rejected"
)

// Loan is the core entity representing a loan application.
type Loan struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	MRP           int64      `json:"mrp"`
	DP            int64      `json:"dp"`
	VehicleYear   int        `json:"vehicle_year"`
	PoliceNumber  string     `json:"police_number"`
	MachineNumber string     `json:"machine_number"`
	Status        LoanStatus `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
