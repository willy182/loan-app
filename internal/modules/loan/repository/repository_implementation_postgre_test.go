//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/willy/loan-app/internal/modules/loan/domain"
	"github.com/willy/loan-app/internal/modules/loan/repository"
)

// setupTestDB starts a PostgreSQL container, runs migrations, and returns a pool.
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:14-bullseye"),
		postgres.WithDatabase("loan_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	// Run migration
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS loans (
			id          UUID PRIMARY KEY,
			user_id     VARCHAR(255) NOT NULL,
			mrp         BIGINT NOT NULL,
			dp          BIGINT NOT NULL,
			vehicle_year INTEGER NOT NULL,
			police_number  VARCHAR(50) NOT NULL,
			machine_number VARCHAR(255) NOT NULL,
			status      VARCHAR(20) NOT NULL DEFAULT 'submitted',
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(user_id, police_number)
		)
	`)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return pool, cleanup
}

func TestPostgreLoanRepository_CreateAndFindByUserID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgreLoanRepository(pool)
	ctx := context.Background()

	loan := &domain.Loan{
		ID:            uuid.New().String(),
		UserID:        "Bruce",
		MRP:           100000000,
		DP:            20000000,
		VehicleYear:   2018,
		PoliceNumber:  "B 1234 BYE",
		MachineNumber: "SDR72V25000W201",
		Status:        domain.LoanStatusSubmitted,
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)
	assert.NotZero(t, loan.CreatedAt)
	assert.NotZero(t, loan.UpdatedAt)

	// Find by UserID
	loans, err := repo.FindByUserID(ctx, "Bruce")
	require.NoError(t, err)
	require.Len(t, loans, 1)
	assert.Equal(t, loan.ID, loans[0].ID)
	assert.Equal(t, domain.LoanStatusSubmitted, loans[0].Status)
	assert.Equal(t, int64(100000000), loans[0].MRP)
}

func TestPostgreLoanRepository_FindByUserIDAndPoliceNumber(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgreLoanRepository(pool)
	ctx := context.Background()

	loan := &domain.Loan{
		ID:            uuid.New().String(),
		UserID:        "Bruce",
		MRP:           100000000,
		DP:            20000000,
		VehicleYear:   2018,
		PoliceNumber:  "B 1234 BYE",
		MachineNumber: "SDR72V25000W201",
		Status:        domain.LoanStatusSubmitted,
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)

	// Find by composite key - found
	found, err := repo.FindByUserIDAndPoliceNumber(ctx, "Bruce", "B 1234 BYE")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, loan.ID, found.ID)
	assert.Equal(t, domain.LoanStatusSubmitted, found.Status)

	// Find by composite key - not found
	notFound, err := repo.FindByUserIDAndPoliceNumber(ctx, "Bruce", "B 9999 XX")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestPostgreLoanRepository_UpdateStatus(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgreLoanRepository(pool)
	ctx := context.Background()

	loan := &domain.Loan{
		ID:            uuid.New().String(),
		UserID:        "Bruce",
		MRP:           100000000,
		DP:            20000000,
		VehicleYear:   2018,
		PoliceNumber:  "B 1234 BYE",
		MachineNumber: "SDR72V25000W201",
		Status:        domain.LoanStatusSubmitted,
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)

	// Update status to approved
	err = repo.UpdateStatus(ctx, "Bruce", "B 1234 BYE", domain.LoanStatusApproved)
	require.NoError(t, err)

	// Verify
	updated, err := repo.FindByUserIDAndPoliceNumber(ctx, "Bruce", "B 1234 BYE")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, domain.LoanStatusApproved, updated.Status)
}

func TestPostgreLoanRepository_FindByUserID_MultipleLoans(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgreLoanRepository(pool)
	ctx := context.Background()

	// Create multiple loans for the same user
	loan1 := &domain.Loan{
		ID:            uuid.New().String(),
		UserID:        "Bruce",
		MRP:           100000000,
		DP:            20000000,
		VehicleYear:   2018,
		PoliceNumber:  "B 1234 BYE",
		MachineNumber: "SDR72V25000W201",
		Status:        domain.LoanStatusSubmitted,
	}
	loan2 := &domain.Loan{
		ID:            uuid.New().String(),
		UserID:        "Bruce",
		MRP:           200000000,
		DP:            50000000,
		VehicleYear:   2020,
		PoliceNumber:  "D 5678 CD",
		MachineNumber: "XYZ123",
		Status:        domain.LoanStatusApproved,
	}

	require.NoError(t, repo.Create(ctx, loan1))
	require.NoError(t, repo.Create(ctx, loan2))

	loans, err := repo.FindByUserID(ctx, "Bruce")
	require.NoError(t, err)
	require.Len(t, loans, 2)
}

func TestPostgreLoanRepository_FindByUserID_EmptyResult(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgreLoanRepository(pool)
	ctx := context.Background()

	loans, err := repo.FindByUserID(ctx, "Nonexistent")
	require.NoError(t, err)
	assert.Empty(t, loans)
}
