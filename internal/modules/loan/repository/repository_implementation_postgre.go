package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/willy/loan-app/internal/modules/loan/domain"
)

// PgxPool is an interface that pgxpool.Pool satisfies, allowing mocking in tests.
type PgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// PostgreLoanRepository implements LoanRepository using PostgreSQL.
type PostgreLoanRepository struct {
	pool PgxPool
}

// NewPostgreLoanRepository creates a new PostgreSQL-backed loan repository.
func NewPostgreLoanRepository(pool PgxPool) *PostgreLoanRepository {
	return &PostgreLoanRepository{pool: pool}
}

// Create inserts a new loan into the database.
func (r *PostgreLoanRepository) Create(ctx context.Context, loan *domain.Loan) error {
	query := `
		INSERT INTO loans (id, user_id, mrp, dp, vehicle_year, police_number, machine_number, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	now := time.Now()
	_, err := r.pool.Exec(ctx, query,
		loan.ID,
		loan.UserID,
		loan.MRP,
		loan.DP,
		loan.VehicleYear,
		loan.PoliceNumber,
		loan.MachineNumber,
		string(loan.Status),
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("inserting loan: %w", err)
	}

	loan.CreatedAt = now
	loan.UpdatedAt = now
	return nil
}

// UpdateStatus changes the status of a loan identified by user_id and police_number.
func (r *PostgreLoanRepository) UpdateStatus(ctx context.Context, userID, policeNumber string, status domain.LoanStatus) error {
	query := `
		UPDATE loans SET status = $1, updated_at = $2
		WHERE user_id = $3 AND police_number = $4
	`
	ct, err := r.pool.Exec(ctx, query, string(status), time.Now(), userID, policeNumber)
	if err != nil {
		return fmt.Errorf("updating loan status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("loan not found for user %s and police %s", userID, policeNumber)
	}
	return nil
}

// FindByUserIDAndPoliceNumber retrieves a single loan by its composite key.
func (r *PostgreLoanRepository) FindByUserIDAndPoliceNumber(ctx context.Context, userID, policeNumber string) (*domain.Loan, error) {
	query := `
		SELECT id, user_id, mrp, dp, vehicle_year, police_number, machine_number, status, created_at, updated_at
		FROM loans
		WHERE user_id = $1 AND police_number = $2
	`
	loan := &domain.Loan{}
	err := r.pool.QueryRow(ctx, query, userID, policeNumber).Scan(
		&loan.ID,
		&loan.UserID,
		&loan.MRP,
		&loan.DP,
		&loan.VehicleYear,
		&loan.PoliceNumber,
		&loan.MachineNumber,
		&loan.Status,
		&loan.CreatedAt,
		&loan.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying loan: %w", err)
	}
	return loan, nil
}

// FindByUserID retrieves all loans belonging to a user.
func (r *PostgreLoanRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Loan, error) {
	query := `
		SELECT id, user_id, mrp, dp, vehicle_year, police_number, machine_number, status, created_at, updated_at
		FROM loans
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying loans for user %s: %w", userID, err)
	}
	defer rows.Close()

	var loans []*domain.Loan
	for rows.Next() {
		loan := &domain.Loan{}
		if err := rows.Scan(
			&loan.ID,
			&loan.UserID,
			&loan.MRP,
			&loan.DP,
			&loan.VehicleYear,
			&loan.PoliceNumber,
			&loan.MachineNumber,
			&loan.Status,
			&loan.CreatedAt,
			&loan.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning loan row: %w", err)
		}
		loans = append(loans, loan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating loan rows: %w", err)
	}

	return loans, nil
}
