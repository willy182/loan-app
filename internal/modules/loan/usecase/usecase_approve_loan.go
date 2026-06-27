package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/willy/loan-app/internal/modules/loan/domain"
)

// ApproveLoan validates the request, finds the loan by user and police number,
// updates its status to "approved", and returns a success response.
func (uc *loanUseCase) ApproveLoan(ctx context.Context, req domain.ApproveLoanRequest) (*domain.ApproveLoanSuccessResponse, error) {
	if err := validateApproveLoan(req); err != nil {
		return nil, fmt.Errorf("validating request: %w", err)
	}

	loan, err := uc.repo.FindByUserIDAndPoliceNumber(ctx, req.UserID, req.PoliceNumber)
	if err != nil {
		return nil, fmt.Errorf("finding loan: %w", err)
	}
	if loan == nil {
		return nil, fmt.Errorf("%w", domain.ErrLoanNotFound)
	}

	if err := uc.repo.UpdateStatus(ctx, req.UserID, req.PoliceNumber, domain.LoanStatusApproved); err != nil {
		return nil, fmt.Errorf("updating loan status: %w", err)
	}

	return &domain.ApproveLoanSuccessResponse{
		UserID:       req.UserID,
		PoliceNumber: req.PoliceNumber,
		Message:      "Loan updated successfully.",
	}, nil
}

func validateApproveLoan(req domain.ApproveLoanRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.PoliceNumber == "" {
		return fmt.Errorf("police_number is required")
	}
	return nil
}

// IsLoanNotFound checks if the error wraps domain.ErrLoanNotFound.
func IsLoanNotFound(err error) bool {
	return errors.Is(err, domain.ErrLoanNotFound)
}
