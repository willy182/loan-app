package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/willy/loan-app/internal/modules/loan/domain"
	"github.com/willy/loan-app/internal/modules/loan/usecase"
)

// LoanHandler handles HTTP requests for loan operations.
type LoanHandler struct {
	useCase usecase.LoanUseCase
	logger  *slog.Logger
}

// NewLoanHandler creates a new LoanHandler with the given use case.
func NewLoanHandler(uc usecase.LoanUseCase) *LoanHandler {
	return &LoanHandler{
		useCase: uc,
		logger:  slog.Default(),
	}
}

// RequestLoan handles POST /api/loans/request
func (h *LoanHandler) RequestLoan(w http.ResponseWriter, r *http.Request) {
	var req domain.RequestLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	resp, err := h.useCase.RequestLoan(r.Context(), req)
	if err != nil {
		// Check for known error types
		if isValidationError(err) {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}
		h.logger.ErrorContext(r.Context(), "request loan failed", "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// ApproveLoan handles POST /api/loans/approve
func (h *LoanHandler) ApproveLoan(w http.ResponseWriter, r *http.Request) {
	var req domain.ApproveLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	resp, err := h.useCase.ApproveLoan(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrLoanNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ApproveLoanErrorResponse{
				Error:            "loan_not_found",
				ErrorDescription: "Loan not Found",
			})
			return
		}

		h.logger.ErrorContext(r.Context(), "approve loan failed", "error", err)
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// isValidationError checks if the error originates from input validation.
// Errors from the use case layer are prefixed with "validating" when they
// come from input validation failures.
func isValidationError(err error) bool {
	return strings.Contains(err.Error(), "validating")
}

// isLoanNotFoundError checks if the error wraps domain.ErrLoanNotFound.
// func isLoanNotFoundError(err error) bool {
// 	return errors.Is(err, domain.ErrLoanNotFound)
// }

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
