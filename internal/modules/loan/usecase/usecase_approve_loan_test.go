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

func TestApproveLoan_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockRepository)
	uc := usecase.NewLoanUseCase(repo)

	existingLoan := &domain.Loan{
		ID:            "loan-uuid-1",
		UserID:        "Bruce",
		MRP:           100000000,
		DP:            20000000,
		VehicleYear:   2018,
		PoliceNumber:  "B 1234 BYE",
		MachineNumber: "SDR72V25000W201",
		Status:        domain.LoanStatusSubmitted,
	}

	req := domain.ApproveLoanRequest{
		UserID:       "Bruce",
		PoliceNumber: "B 1234 BYE",
	}

	repo.On("FindByUserIDAndPoliceNumber", mock.Anything, "Bruce", "B 1234 BYE").
		Return(existingLoan, nil).Once()
	repo.On("UpdateStatus", mock.Anything, "Bruce", "B 1234 BYE", domain.LoanStatusApproved).
		Return(nil).Once()

	resp, err := uc.ApproveLoan(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "Bruce", resp.UserID)
	assert.Equal(t, "B 1234 BYE", resp.PoliceNumber)
	assert.Equal(t, "Loan updated successfully.", resp.Message)

	repo.AssertExpectations(t)
}

func TestApproveLoan_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     domain.ApproveLoanRequest
		wantErr string
	}{
		{
			name: "empty user_id",
			req: domain.ApproveLoanRequest{
				UserID:       "",
				PoliceNumber: "B 1234 BYE",
			},
			wantErr: "user_id is required",
		},
		{
			name: "empty police_number",
			req: domain.ApproveLoanRequest{
				UserID:       "Bruce",
				PoliceNumber: "",
			},
			wantErr: "police_number is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepository)
			uc := usecase.NewLoanUseCase(repo)

			_, err := uc.ApproveLoan(context.Background(), tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)

			repo.AssertNotCalled(t, "FindByUserIDAndPoliceNumber")
			repo.AssertNotCalled(t, "UpdateStatus")
		})
	}
}

func TestApproveLoan_LoanNotFound(t *testing.T) {
	t.Parallel()

	repo := new(mockRepository)
	uc := usecase.NewLoanUseCase(repo)

	req := domain.ApproveLoanRequest{
		UserID:       "Bruce",
		PoliceNumber: "B 1234 BYE",
	}

	repo.On("FindByUserIDAndPoliceNumber", mock.Anything, "Bruce", "B 1234 BYE").
		Return(nil, nil).Once()

	_, err := uc.ApproveLoan(context.Background(), req)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrLoanNotFound)

	repo.AssertExpectations(t)
}

func TestApproveLoan_RepositoryError(t *testing.T) {
	t.Parallel()

	repo := new(mockRepository)
	uc := usecase.NewLoanUseCase(repo)

	req := domain.ApproveLoanRequest{
		UserID:       "Bruce",
		PoliceNumber: "B 1234 BYE",
	}

	expectedErr := errors.New("connection timeout")
	repo.On("FindByUserIDAndPoliceNumber", mock.Anything, "Bruce", "B 1234 BYE").
		Return(nil, expectedErr).Once()

	_, err := uc.ApproveLoan(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "finding loan")

	repo.AssertExpectations(t)
}

func TestApproveLoan_UpdateStatusError(t *testing.T) {
	t.Parallel()

	repo := new(mockRepository)
	uc := usecase.NewLoanUseCase(repo)

	existingLoan := &domain.Loan{
		ID:            "loan-uuid-1",
		UserID:        "Bruce",
		MRP:           100000000,
		DP:            20000000,
		VehicleYear:   2018,
		PoliceNumber:  "B 1234 BYE",
		MachineNumber: "SDR72V25000W201",
		Status:        domain.LoanStatusSubmitted,
	}

	req := domain.ApproveLoanRequest{
		UserID:       "Bruce",
		PoliceNumber: "B 1234 BYE",
	}

	repo.On("FindByUserIDAndPoliceNumber", mock.Anything, "Bruce", "B 1234 BYE").
		Return(existingLoan, nil).Once()
	repo.On("UpdateStatus", mock.Anything, "Bruce", "B 1234 BYE", domain.LoanStatusApproved).
		Return(errors.New("update failed")).Once()

	_, err := uc.ApproveLoan(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "updating loan status")

	repo.AssertExpectations(t)
}
