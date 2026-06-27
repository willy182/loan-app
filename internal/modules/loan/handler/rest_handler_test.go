package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/willy/loan-app/internal/modules/loan/domain"
	"github.com/willy/loan-app/internal/modules/loan/handler"
)

// mockUseCase implements usecase.LoanUseCase for testing handlers.
type mockUseCase struct {
	mock.Mock
}

func (m *mockUseCase) RequestLoan(ctx context.Context, req domain.RequestLoanRequest) (*domain.RequestLoanResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RequestLoanResponse), args.Error(1)
}

func (m *mockUseCase) ApproveLoan(ctx context.Context, req domain.ApproveLoanRequest) (*domain.ApproveLoanSuccessResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ApproveLoanSuccessResponse), args.Error(1)
}

func setupRouter(uc *mockUseCase) chi.Router {
	r := chi.NewRouter()
	h := handler.NewLoanHandler(uc)
	r.Post("/api/loans/request", h.RequestLoan)
	r.Post("/api/loans/approve", h.ApproveLoan)
	return r
}

func TestRequestLoanHandler_Success(t *testing.T) {
	t.Parallel()

	uc := new(mockUseCase)
	router := setupRouter(uc)

	reqBody := `{
		"user_id": "Bruce",
		"mrp": 100000000,
		"dp": 20000000,
		"vehicle_year": 2018,
		"police_number": "B 1234 BYE",
		"machine_number": "SDR72V25000W201"
	}`

	expectedResp := &domain.RequestLoanResponse{
		UserID: "Bruce",
		Loans: []domain.LoanItem{
			{
				MRP:           100000000,
				DP:            20000000,
				VehicleYear:   2018,
				PoliceNumber:  "B 1234 BYE",
				MachineNumber: "SDR72V25000W201",
				Status:        domain.LoanStatusSubmitted,
			},
		},
	}

	uc.On("RequestLoan", mock.Anything, mock.MatchedBy(func(req domain.RequestLoanRequest) bool {
		return req.UserID == "Bruce" &&
			req.MRP == 100000000 &&
			req.DP == 20000000 &&
			req.VehicleYear == 2018 &&
			req.PoliceNumber == "B 1234 BYE" &&
			req.MachineNumber == "SDR72V25000W201"
	})).Return(expectedResp, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/loans/request", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp domain.RequestLoanResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "Bruce", resp.UserID)
	assert.Len(t, resp.Loans, 1)
	assert.Equal(t, domain.LoanStatusSubmitted, resp.Loans[0].Status)

	uc.AssertExpectations(t)
}

func TestRequestLoanHandler_InvalidJSON(t *testing.T) {
	t.Parallel()

	uc := new(mockUseCase)
	router := setupRouter(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/loans/request", strings.NewReader(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "invalid request")

	uc.AssertNotCalled(t, "RequestLoan")
}

func TestRequestLoanHandler_ValidationError(t *testing.T) {
	t.Parallel()

	uc := new(mockUseCase)
	router := setupRouter(uc)

	reqBody := `{"user_id": "", "mrp": 100000000, "dp": 20000000, "vehicle_year": 2018, "police_number": "B 1234 BYE", "machine_number": "SDR72V25000W201"}`

	uc.On("RequestLoan", mock.Anything, mock.Anything).
		Return(nil, errors.New("validating request: user_id is required")).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/loans/request", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	uc.AssertExpectations(t)
}

func TestApproveLoanHandler_Success(t *testing.T) {
	t.Parallel()

	uc := new(mockUseCase)
	router := setupRouter(uc)

	reqBody := `{"user_id": "Bruce", "police_number": "B 1234 BYE"}`

	expectedResp := &domain.ApproveLoanSuccessResponse{
		UserID:       "Bruce",
		PoliceNumber: "B 1234 BYE",
		Message:      "Loan updated successfully.",
	}

	uc.On("ApproveLoan", mock.Anything, mock.MatchedBy(func(req domain.ApproveLoanRequest) bool {
		return req.UserID == "Bruce" && req.PoliceNumber == "B 1234 BYE"
	})).Return(expectedResp, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/loans/approve", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp domain.ApproveLoanSuccessResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "Bruce", resp.UserID)
	assert.Equal(t, "B 1234 BYE", resp.PoliceNumber)
	assert.Equal(t, "Loan updated successfully.", resp.Message)

	uc.AssertExpectations(t)
}

func TestApproveLoanHandler_NotFound(t *testing.T) {
	t.Parallel()

	uc := new(mockUseCase)
	router := setupRouter(uc)

	reqBody := `{"user_id": "Bruce", "police_number": "B 9999 XX"}`

	uc.On("ApproveLoan", mock.Anything, mock.Anything).
		Return(nil, domain.ErrLoanNotFound).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/loans/approve", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)

	var errResp domain.ApproveLoanErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)

	assert.Equal(t, "loan_not_found", errResp.Error)
	assert.Equal(t, "Loan not Found", errResp.ErrorDescription)

	uc.AssertExpectations(t)
}

func TestApproveLoanHandler_InvalidJSON(t *testing.T) {
	t.Parallel()

	uc := new(mockUseCase)
	router := setupRouter(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/loans/approve", strings.NewReader(`{broken`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "invalid request")

	uc.AssertNotCalled(t, "ApproveLoan")
}

func TestRequestLoanHandler_InternalError(t *testing.T) {
	t.Parallel()

	uc := new(mockUseCase)
	router := setupRouter(uc)

	reqBody := `{
		"user_id": "Bruce",
		"mrp": 100000000,
		"dp": 20000000,
		"vehicle_year": 2018,
		"police_number": "B 1234 BYE",
		"machine_number": "SDR72V25000W201"
	}`

	uc.On("RequestLoan", mock.Anything, mock.Anything).
		Return(nil, errors.New("internal error")).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/loans/request", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	uc.AssertExpectations(t)
}
