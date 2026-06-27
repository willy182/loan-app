package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/willy/loan-app/internal/modules/loan/domain"
	"github.com/willy/loan-app/internal/modules/loan/repository"
)

// setupGormMock creates a sqlmock-backed GORM DB for unit testing.
func setupGormMock(t *testing.T) (sqlmock.Sqlmock, *gorm.DB) {
	t.Helper()

	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	return mock, db
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestGormRepo_Create_Success(t *testing.T) {
	t.Parallel()

	mock, db := setupGormMock(t)
	repo := repository.NewGormLoanRepository(db)
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

	// GORM INSERT uses a transaction with Query and RETURNING "id"
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "loans"`)).
		WithArgs("Bruce", int64(100000000), int64(20000000), 2018,
			"B 1234 BYE", "SDR72V25000W201", "submitted",
			sqlmock.AnyArg(), sqlmock.AnyArg(), loanID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(loanID))
	mock.ExpectCommit()

	err := repo.Create(ctx, loan)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGormRepo_Create_DBError(t *testing.T) {
	t.Parallel()

	mock, db := setupGormMock(t)
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

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "loans"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrInvalidData)
	mock.ExpectRollback()

	err := repo.Create(ctx, loan)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inserting loan")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// UpdateStatus
// ---------------------------------------------------------------------------

func TestGormRepo_UpdateStatus_Success(t *testing.T) {
	t.Parallel()

	mock, db := setupGormMock(t)
	repo := repository.NewGormLoanRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "loans" SET "status"=$1,"updated_at"=$2 WHERE user_id = $3 AND police_number = $4`)).
		WithArgs("approved", sqlmock.AnyArg(), "Bruce", "B 1234 BYE").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateStatus(ctx, "Bruce", "B 1234 BYE", domain.LoanStatusApproved)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGormRepo_UpdateStatus_NotFound(t *testing.T) {
	t.Parallel()

	mock, db := setupGormMock(t)
	repo := repository.NewGormLoanRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "loans" SET "status"=$1,"updated_at"=$2 WHERE user_id = $3 AND police_number = $4`)).
		WithArgs("approved", sqlmock.AnyArg(), "Bruce", "B 9999 XX").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.UpdateStatus(ctx, "Bruce", "B 9999 XX", domain.LoanStatusApproved)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loan not found")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGormRepo_UpdateStatus_DBError(t *testing.T) {
	t.Parallel()

	mock, db := setupGormMock(t)
	repo := repository.NewGormLoanRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "loans" SET "status"=$1,"updated_at"=$2 WHERE user_id = $3 AND police_number = $4`)).
		WithArgs("approved", sqlmock.AnyArg(), "Bruce", "B 1234 BYE").
		WillReturnError(gorm.ErrInvalidTransaction)
	mock.ExpectRollback()

	err := repo.UpdateStatus(ctx, "Bruce", "B 1234 BYE", domain.LoanStatusApproved)
	require.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// FindByUserIDAndPoliceNumber
// ---------------------------------------------------------------------------

func TestGormRepo_FindByUserIDAndPoliceNumber_Found(t *testing.T) {
	t.Parallel()

	mock, db := setupGormMock(t)
	repo := repository.NewGormLoanRepository(db)
	ctx := context.Background()

	now := time.Now()
	loanID := uuid.New().String()

	rows := sqlmock.NewRows([]string{"id", "user_id", "mrp", "dp", "vehicle_year",
		"police_number", "machine_number", "status", "created_at", "updated_at"}).
		AddRow(loanID, "Bruce", int64(100000000), int64(20000000), 2018,
			"B 1234 BYE", "SDR72V25000W201", "submitted", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "loans" WHERE user_id = $1 AND police_number = $2 ORDER BY "loans"."id" LIMIT $3`)).
		WithArgs("Bruce", "B 1234 BYE", 1).
		WillReturnRows(rows)

	loan, err := repo.FindByUserIDAndPoliceNumber(ctx, "Bruce", "B 1234 BYE")
	require.NoError(t, err)
	require.NotNil(t, loan)

	assert.Equal(t, loanID, loan.ID)
	assert.Equal(t, "Bruce", loan.UserID)
	assert.Equal(t, int64(100000000), loan.MRP)
	assert.Equal(t, domain.LoanStatusSubmitted, loan.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGormRepo_FindByUserIDAndPoliceNumber_NotFound(t *testing.T) {
	t.Parallel()

	mock, db := setupGormMock(t)
	repo := repository.NewGormLoanRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "loans" WHERE user_id = $1 AND police_number = $2 ORDER BY "loans"."id" LIMIT $3`)).
		WithArgs("Bruce", "B 9999 XX", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	loan, err := repo.FindByUserIDAndPoliceNumber(ctx, "Bruce", "B 9999 XX")
	require.NoError(t, err)
	assert.Nil(t, loan)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// FindByUserID
// ---------------------------------------------------------------------------

func TestGormRepo_FindByUserID_Success(t *testing.T) {
	t.Parallel()

	mock, db := setupGormMock(t)
	repo := repository.NewGormLoanRepository(db)
	ctx := context.Background()

	now := time.Now()
	loanID1 := uuid.New().String()
	loanID2 := uuid.New().String()

	rows := sqlmock.NewRows([]string{"id", "user_id", "mrp", "dp", "vehicle_year",
		"police_number", "machine_number", "status", "created_at", "updated_at"}).
		AddRow(loanID1, "Bruce", int64(100000000), int64(20000000), 2018,
			"B 1234 BYE", "SDR72V25000W201", "submitted", now, now).
		AddRow(loanID2, "Bruce", int64(200000000), int64(50000000), 2020,
			"D 5678 CD", "XYZ123", "approved", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "loans" WHERE user_id = $1 ORDER BY created_at DESC`)).
		WithArgs("Bruce").
		WillReturnRows(rows)

	loans, err := repo.FindByUserID(ctx, "Bruce")
	require.NoError(t, err)
	require.Len(t, loans, 2)

	assert.Equal(t, loanID1, loans[0].ID)
	assert.Equal(t, loanID2, loans[1].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGormRepo_FindByUserID_EmptyResult(t *testing.T) {
	t.Parallel()

	mock, db := setupGormMock(t)
	repo := repository.NewGormLoanRepository(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "user_id", "mrp", "dp", "vehicle_year",
		"police_number", "machine_number", "status", "created_at", "updated_at"})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "loans" WHERE user_id = $1 ORDER BY created_at DESC`)).
		WithArgs("Unknown").
		WillReturnRows(rows)

	loans, err := repo.FindByUserID(ctx, "Unknown")
	require.NoError(t, err)
	assert.Empty(t, loans)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGormRepo_FindByUserID_DBError(t *testing.T) {
	t.Parallel()

	mock, db := setupGormMock(t)
	repo := repository.NewGormLoanRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "loans" WHERE user_id = $1 ORDER BY created_at DESC`)).
		WithArgs("Bruce").
		WillReturnError(gorm.ErrInvalidDB)

	_, err := repo.FindByUserID(ctx, "Bruce")
	require.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
