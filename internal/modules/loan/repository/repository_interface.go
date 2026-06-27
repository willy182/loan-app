package repository

import (
	"context"

	"github.com/willy/loan-app/internal/modules/loan/domain"
)

// LoanRepository defines the persistence contract for loan data.
type LoanRepository interface {
	// Create inserts a new loan into the database.
	Create(ctx context.Context, loan *domain.Loan) error

	// UpdateStatus changes the status of a loan identified by user_id and police_number.
	UpdateStatus(ctx context.Context, userID, policeNumber string, status domain.LoanStatus) error

	// FindByUserIDAndPoliceNumber retrieves a single loan by its composite key.
	FindByUserIDAndPoliceNumber(ctx context.Context, userID, policeNumber string) (*domain.Loan, error)

	// FindByUserID retrieves all loans belonging to a user.
	FindByUserID(ctx context.Context, userID string) ([]*domain.Loan, error)
}
