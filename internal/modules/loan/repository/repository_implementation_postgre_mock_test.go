package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/willy/loan-app/internal/modules/loan/domain"
	"github.com/willy/loan-app/internal/modules/loan/repository"
)

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestPgxRepo_Create_Success(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := repository.NewPostgreLoanRepository(mockPool)
	ctx := context.Background()

	loanID := uuid.New().String()
	loan := &domain.Loan{
		ID:            loanID,
		UserID:        "Bruce",
		MRP:           100000000,
		DP:            20000000,
		VehicleYear:   2018,
		PoliceNumber:  "B 1234 BYE",
		MachineNumber: "SDR72V25000W201",
		Status:        domain.LoanStatusSubmitted,
	}

	mockPool.ExpectExec(`INSERT INTO loans`).
		WithArgs(loanID, "Bruce", int64(100000000), int64(20000000), 2018,
			"B 1234 BYE", "SDR72V25000W201", "submitted",
			pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(ctx, loan)
	require.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestPgxRepo_Create_DBError(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := repository.NewPostgreLoanRepository(mockPool)
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

	mockPool.ExpectExec(`INSERT INTO loans`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(errors.New("connection error"))

	err = repo.Create(ctx, loan)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inserting loan")

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// UpdateStatus
// ---------------------------------------------------------------------------

func TestPgxRepo_UpdateStatus_Success(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := repository.NewPostgreLoanRepository(mockPool)
	ctx := context.Background()

	mockPool.ExpectExec(`UPDATE loans SET status`).
		WithArgs("approved", pgxmock.AnyArg(), "Bruce", "B 1234 BYE").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.UpdateStatus(ctx, "Bruce", "B 1234 BYE", domain.LoanStatusApproved)
	require.NoError(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestPgxRepo_UpdateStatus_NotFound(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := repository.NewPostgreLoanRepository(mockPool)
	ctx := context.Background()

	mockPool.ExpectExec(`UPDATE loans SET status`).
		WithArgs("approved", pgxmock.AnyArg(), "Bruce", "B 9999 XX").
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err = repo.UpdateStatus(ctx, "Bruce", "B 9999 XX", domain.LoanStatusApproved)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loan not found")

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestPgxRepo_UpdateStatus_DBError(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := repository.NewPostgreLoanRepository(mockPool)
	ctx := context.Background()

	mockPool.ExpectExec(`UPDATE loans SET status`).
		WithArgs("approved", pgxmock.AnyArg(), "Bruce", "B 1234 BYE").
		WillReturnError(errors.New("update failed"))

	err = repo.UpdateStatus(ctx, "Bruce", "B 1234 BYE", domain.LoanStatusApproved)
	require.Error(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// FindByUserIDAndPoliceNumber
// ---------------------------------------------------------------------------

func TestPgxRepo_FindByUserIDAndPoliceNumber_Found(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := repository.NewPostgreLoanRepository(mockPool)
	ctx := context.Background()

	loanID := uuid.New().String()

	rows := mockPool.NewRows([]string{"id", "user_id", "mrp", "dp", "vehicle_year",
		"police_number", "machine_number", "status", "created_at", "updated_at"}).
		AddRow(loanID, "Bruce", int64(100000000), int64(20000000), 2018,
			"B 1234 BYE", "SDR72V25000W201", domain.LoanStatus("submitted"), time.Now(), time.Now())

	mockPool.ExpectQuery(`SELECT id, user_id, mrp, dp, vehicle_year, police_number, machine_number, status, created_at, updated_at FROM loans WHERE user_id = \$1 AND police_number = \$2`).
		WithArgs("Bruce", "B 1234 BYE").
		WillReturnRows(rows)

	found, err := repo.FindByUserIDAndPoliceNumber(ctx, "Bruce", "B 1234 BYE")
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, loanID, found.ID)
	assert.Equal(t, "Bruce", found.UserID)
	assert.Equal(t, int64(100000000), found.MRP)
	assert.Equal(t, domain.LoanStatusSubmitted, found.Status)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestPgxRepo_FindByUserIDAndPoliceNumber_NotFound(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := repository.NewPostgreLoanRepository(mockPool)
	ctx := context.Background()

	mockPool.ExpectQuery(`SELECT id, user_id, mrp, dp, vehicle_year, police_number, machine_number, status, created_at, updated_at FROM loans WHERE user_id = \$1 AND police_number = \$2`).
		WithArgs("Bruce", "B 9999 XX").
		WillReturnError(pgx.ErrNoRows)

	loan, err := repo.FindByUserIDAndPoliceNumber(ctx, "Bruce", "B 9999 XX")
	require.NoError(t, err)
	assert.Nil(t, loan)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// FindByUserID
// ---------------------------------------------------------------------------

func TestPgxRepo_FindByUserID_Success(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := repository.NewPostgreLoanRepository(mockPool)
	ctx := context.Background()

	loanID1 := uuid.New().String()
	loanID2 := uuid.New().String()

	rows := mockPool.NewRows([]string{"id", "user_id", "mrp", "dp", "vehicle_year",
		"police_number", "machine_number", "status", "created_at", "updated_at"}).
		AddRow(loanID1, "Bruce", int64(100000000), int64(20000000), 2018,
			"B 1234 BYE", "SDR72V25000W201", domain.LoanStatus("submitted"), time.Now(), time.Now()).
		AddRow(loanID2, "Bruce", int64(200000000), int64(50000000), 2020,
			"D 5678 CD", "XYZ123", domain.LoanStatus("approved"), time.Now(), time.Now())

	mockPool.ExpectQuery(`SELECT id, user_id, mrp, dp, vehicle_year, police_number, machine_number, status, created_at, updated_at FROM loans WHERE user_id = \$1 ORDER BY created_at DESC`).
		WithArgs("Bruce").
		WillReturnRows(rows)

	loans, err := repo.FindByUserID(ctx, "Bruce")
	require.NoError(t, err)
	require.Len(t, loans, 2)

	assert.Equal(t, loanID1, loans[0].ID)
	assert.Equal(t, loanID2, loans[1].ID)
	assert.Equal(t, domain.LoanStatusSubmitted, loans[0].Status)
	assert.Equal(t, domain.LoanStatusApproved, loans[1].Status)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestPgxRepo_FindByUserID_EmptyResult(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := repository.NewPostgreLoanRepository(mockPool)
	ctx := context.Background()

	rows := mockPool.NewRows([]string{"id", "user_id", "mrp", "dp", "vehicle_year",
		"police_number", "machine_number", "status", "created_at", "updated_at"})

	mockPool.ExpectQuery(`SELECT id, user_id, mrp, dp, vehicle_year, police_number, machine_number, status, created_at, updated_at FROM loans WHERE user_id = \$1 ORDER BY created_at DESC`).
		WithArgs("Unknown").
		WillReturnRows(rows)

	loans, err := repo.FindByUserID(ctx, "Unknown")
	require.NoError(t, err)
	assert.Empty(t, loans)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestPgxRepo_FindByUserID_DBError(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := repository.NewPostgreLoanRepository(mockPool)
	ctx := context.Background()

	mockPool.ExpectQuery(`SELECT id, user_id, mrp, dp, vehicle_year, police_number, machine_number, status, created_at, updated_at FROM loans WHERE user_id = \$1 ORDER BY created_at DESC`).
		WithArgs("Bruce").
		WillReturnError(errors.New("query failed"))

	_, err = repo.FindByUserID(ctx, "Bruce")
	require.Error(t, err)

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
