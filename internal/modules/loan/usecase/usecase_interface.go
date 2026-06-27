package usecase

import (
	"context"

	"github.com/willy/loan-app/internal/modules/loan/domain"
)

// LoanUseCase defines the application-level operations for loan processing.
type LoanUseCase interface {
	// RequestLoan initiates a new loan application. Returns the created loan details.
	RequestLoan(ctx context.Context, req domain.RequestLoanRequest) (*domain.RequestLoanResponse, error)

	// ApproveLoan approves a submitted loan application.
	ApproveLoan(ctx context.Context, req domain.ApproveLoanRequest) (*domain.ApproveLoanSuccessResponse, error)
}
