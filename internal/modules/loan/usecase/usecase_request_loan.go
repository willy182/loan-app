package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/willy/loan-app/internal/modules/loan/domain"
	"github.com/willy/loan-app/internal/modules/loan/repository"
)

// loanUseCase implements the LoanUseCase interface.
type loanUseCase struct {
	repo repository.LoanRepository
}

// NewLoanUseCase creates a new loan use case with the given repository.
func NewLoanUseCase(repo repository.LoanRepository) LoanUseCase {
	return &loanUseCase{repo: repo}
}

// RequestLoan validates the input, creates a new loan with "submitted" status,
// and returns the response with all loans for the user.
func (uc *loanUseCase) RequestLoan(ctx context.Context, req domain.RequestLoanRequest) (*domain.RequestLoanResponse, error) {
	if err := validateRequestLoan(req); err != nil {
		return nil, fmt.Errorf("validating request: %w", err)
	}

	// TODO: Check if a loan with the same police number already exists for the user using redis
	// if key redis with the police number is exists
	// return error "loan with this police number already exists for the user"

	loan := &domain.Loan{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		MRP:           req.MRP,
		DP:            req.DP,
		VehicleYear:   req.VehicleYear,
		PoliceNumber:  req.PoliceNumber,
		MachineNumber: req.MachineNumber,
		Status:        domain.LoanStatusSubmitted,
	}

	if err := uc.repo.Create(ctx, loan); err != nil {
		return nil, fmt.Errorf("creating loan: %w", err)
	}

	loans, err := uc.repo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetching user loans: %w", err)
	}

	items := make([]domain.LoanItem, 0, len(loans))
	for _, l := range loans {
		items = append(items, domain.LoanItem{
			MRP:           l.MRP,
			DP:            l.DP,
			VehicleYear:   l.VehicleYear,
			PoliceNumber:  l.PoliceNumber,
			MachineNumber: l.MachineNumber,
			Status:        l.Status,
		})
	}

	return &domain.RequestLoanResponse{
		UserID: req.UserID,
		Loans:  items,
	}, nil
}

func validateRequestLoan(req domain.RequestLoanRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.MRP <= 0 {
		return fmt.Errorf("mrp must be greater than 0")
	}
	if req.DP <= 0 {
		return fmt.Errorf("dp must be greater than 0")
	}
	if req.DP > req.MRP {
		return fmt.Errorf("dp cannot exceed mrp")
	}
	if req.VehicleYear < 1900 || req.VehicleYear > 2031 {
		return fmt.Errorf("vehicle_year must be between 1900 and 2031")
	}
	if req.PoliceNumber == "" {
		return fmt.Errorf("police_number is required")
	}
	if req.MachineNumber == "" {
		return fmt.Errorf("machine_number is required")
	}
	return nil
}
