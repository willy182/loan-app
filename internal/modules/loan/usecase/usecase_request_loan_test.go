package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/willy/loan-app/internal/modules/loan/domain"
	"github.com/willy/loan-app/internal/modules/loan/usecase"
)

// mockRepository implements repository.LoanRepository for testing.
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(ctx context.Context, loan *domain.Loan) error {
	args := m.Called(ctx, loan)
	return args.Error(0)
}

func (m *mockRepository) UpdateStatus(ctx context.Context, userID, policeNumber string, status domain.LoanStatus) error {
	args := m.Called(ctx, userID, policeNumber, status)
	return args.Error(0)
}

func (m *mockRepository) FindByUserIDAndPoliceNumber(ctx context.Context, userID, policeNumber string) (*domain.Loan, error) {
	args := m.Called(ctx, userID, policeNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Loan), args.Error(1)
}

func (m *mockRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Loan, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Loan), args.Error(1)
}

func TestRequestLoan_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockRepository)
	uc := usecase.NewLoanUseCase(repo)

	req := domain.RequestLoanRequest{
		UserID:        "Bruce",
		MRP:           100000000,
		DP:            20000000,
		VehicleYear:   2018,
		PoliceNumber:  "B 1234 BYE",
		MachineNumber: "SDR72V25000W201",
	}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(loan *domain.Loan) bool {
		return loan.UserID == "Bruce" &&
			loan.MRP == 100000000 &&
			loan.DP == 20000000 &&
			loan.VehicleYear == 2018 &&
			loan.PoliceNumber == "B 1234 BYE" &&
			loan.MachineNumber == "SDR72V25000W201" &&
			loan.Status == domain.LoanStatusSubmitted &&
			loan.ID != ""
	})).Return(nil).Once()

	// We also need FindByUserID to be called after Create to build the response
	repo.On("FindByUserID", mock.Anything, "Bruce").Return([]*domain.Loan{
		{
			ID:            "loan-uuid-1",
			UserID:        "Bruce",
			MRP:           100000000,
			DP:            20000000,
			VehicleYear:   2018,
			PoliceNumber:  "B 1234 BYE",
			MachineNumber: "SDR72V25000W201",
			Status:        domain.LoanStatusSubmitted,
		},
	}, nil).Once()

	resp, err := uc.RequestLoan(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "Bruce", resp.UserID)
	assert.Len(t, resp.Loans, 1)
	assert.Equal(t, int64(100000000), resp.Loans[0].MRP)
	assert.Equal(t, int64(20000000), resp.Loans[0].DP)
	assert.Equal(t, 2018, resp.Loans[0].VehicleYear)
	assert.Equal(t, "B 1234 BYE", resp.Loans[0].PoliceNumber)
	assert.Equal(t, "SDR72V25000W201", resp.Loans[0].MachineNumber)
	assert.Equal(t, domain.LoanStatusSubmitted, resp.Loans[0].Status)

	repo.AssertExpectations(t)
}

func TestRequestLoan_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     domain.RequestLoanRequest
		wantErr string
	}{
		{
			name: "empty user_id",
			req: domain.RequestLoanRequest{
				UserID:        "",
				MRP:           100000000,
				DP:            20000000,
				VehicleYear:   2018,
				PoliceNumber:  "B 1234 BYE",
				MachineNumber: "SDR72V25000W201",
			},
			wantErr: "user_id is required",
		},
		{
			name: "zero mrp",
			req: domain.RequestLoanRequest{
				UserID:        "Bruce",
				MRP:           0,
				DP:            20000000,
				VehicleYear:   2018,
				PoliceNumber:  "B 1234 BYE",
				MachineNumber: "SDR72V25000W201",
			},
			wantErr: "mrp must be greater than 0",
		},
		{
			name: "negative mrp",
			req: domain.RequestLoanRequest{
				UserID:        "Bruce",
				MRP:           -100,
				DP:            20000000,
				VehicleYear:   2018,
				PoliceNumber:  "B 1234 BYE",
				MachineNumber: "SDR72V25000W201",
			},
			wantErr: "mrp must be greater than 0",
		},
		{
			name: "zero dp",
			req: domain.RequestLoanRequest{
				UserID:        "Bruce",
				MRP:           100000000,
				DP:            0,
				VehicleYear:   2018,
				PoliceNumber:  "B 1234 BYE",
				MachineNumber: "SDR72V25000W201",
			},
			wantErr: "dp must be greater than 0",
		},
		{
			name: "dp exceeds mrp",
			req: domain.RequestLoanRequest{
				UserID:        "Bruce",
				MRP:           100000000,
				DP:            200000000,
				VehicleYear:   2018,
				PoliceNumber:  "B 1234 BYE",
				MachineNumber: "SDR72V25000W201",
			},
			wantErr: "dp cannot exceed mrp",
		},
		{
			name: "invalid vehicle year - too old",
			req: domain.RequestLoanRequest{
				UserID:        "Bruce",
				MRP:           100000000,
				DP:            20000000,
				VehicleYear:   1899,
				PoliceNumber:  "B 1234 BYE",
				MachineNumber: "SDR72V25000W201",
			},
			wantErr: "vehicle_year must be between 1900 and",
		},
		{
			name: "invalid vehicle year - future",
			req: domain.RequestLoanRequest{
				UserID:        "Bruce",
				MRP:           100000000,
				DP:            20000000,
				VehicleYear:   2051,
				PoliceNumber:  "B 1234 BYE",
				MachineNumber: "SDR72V25000W201",
			},
			wantErr: "vehicle_year must be between 1900 and",
		},
		{
			name: "empty police_number",
			req: domain.RequestLoanRequest{
				UserID:        "Bruce",
				MRP:           100000000,
				DP:            20000000,
				VehicleYear:   2018,
				PoliceNumber:  "",
				MachineNumber: "SDR72V25000W201",
			},
			wantErr: "police_number is required",
		},
		{
			name: "empty machine_number",
			req: domain.RequestLoanRequest{
				UserID:        "Bruce",
				MRP:           100000000,
				DP:            20000000,
				VehicleYear:   2018,
				PoliceNumber:  "B 1234 BYE",
				MachineNumber: "",
			},
			wantErr: "machine_number is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepository)
			uc := usecase.NewLoanUseCase(repo)

			_, err := uc.RequestLoan(context.Background(), tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)

			repo.AssertNotCalled(t, "Create")
		})
	}
}

func TestRequestLoan_RepositoryError(t *testing.T) {
	t.Parallel()

	repo := new(mockRepository)
	uc := usecase.NewLoanUseCase(repo)

	req := domain.RequestLoanRequest{
		UserID:        "Bruce",
		MRP:           100000000,
		DP:            20000000,
		VehicleYear:   2018,
		PoliceNumber:  "B 1234 BYE",
		MachineNumber: "SDR72V25000W201",
	}

	expectedErr := errors.New("connection timeout")
	repo.On("Create", mock.Anything, mock.Anything).Return(expectedErr).Once()

	_, err := uc.RequestLoan(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating loan")

	repo.AssertExpectations(t)
}
