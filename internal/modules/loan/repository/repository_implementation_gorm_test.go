//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/willy/loan-app/internal/modules/loan/domain"
	"github.com/willy/loan-app/internal/modules/loan/repository"
)

// setupGormTestDB starts a PostgreSQL container, runs GORM AutoMigrate, and returns a *gorm.DB.
func setupGormTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:14-bullseye"),
		tcpostgres.WithDatabase("loan_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(gormpg.Open(connStr), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)

	// AutoMigrate
	err = db.AutoMigrate(&repository.GormLoan{})
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return db, cleanup
}

func TestGormLoanRepository_CreateAndFindByUserID(t *testing.T) {
	db, cleanup := setupGormTestDB(t)
	defer cleanup()

	repo := repository.NewGormLoanRepository(db)
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

func TestGormLoanRepository_FindByUserIDAndPoliceNumber(t *testing.T) {
	db, cleanup := setupGormTestDB(t)
	defer cleanup()

	repo := repository.NewGormLoanRepository(db)
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

func TestGormLoanRepository_UpdateStatus(t *testing.T) {
	db, cleanup := setupGormTestDB(t)
	defer cleanup()

	repo := repository.NewGormLoanRepository(db)
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

func TestGormLoanRepository_FindByUserID_MultipleLoans(t *testing.T) {
	db, cleanup := setupGormTestDB(t)
	defer cleanup()

	repo := repository.NewGormLoanRepository(db)
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

func TestGormLoanRepository_FindByUserID_EmptyResult(t *testing.T) {
	db, cleanup := setupGormTestDB(t)
	defer cleanup()

	repo := repository.NewGormLoanRepository(db)
	ctx := context.Background()

	loans, err := repo.FindByUserID(ctx, "Nonexistent")
	require.NoError(t, err)
	assert.Empty(t, loans)
}
