package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/willy/loan-app/internal/modules/loan/domain"
)

// GormLoan is the GORM model for the loans table.
// It maps to the loans table with explicit table name and column mappings.
type GormLoan struct {
	ID            string          `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID        string          `gorm:"column:user_id;not null;index"`
	MRP           int64           `gorm:"column:mrp;not null"`
	DP            int64           `gorm:"column:dp;not null"`
	VehicleYear   int             `gorm:"column:vehicle_year;not null"`
	PoliceNumber  string          `gorm:"column:police_number;not null"`
	MachineNumber string          `gorm:"column:machine_number;not null"`
	Status        domain.LoanStatus `gorm:"column:status;not null;default:'submitted';type:varchar(20)"`
	CreatedAt     time.Time       `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;not null;autoUpdateTime"`
}

// TableName overrides the default table name for GORM.
func (GormLoan) TableName() string {
	return "loans"
}

// toDomain converts a GORM model to a domain entity.
func (g *GormLoan) toDomain() *domain.Loan {
	return &domain.Loan{
		ID:            g.ID,
		UserID:        g.UserID,
		MRP:           g.MRP,
		DP:            g.DP,
		VehicleYear:   g.VehicleYear,
		PoliceNumber:  g.PoliceNumber,
		MachineNumber: g.MachineNumber,
		Status:        g.Status,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
}

// fromDomain converts a domain entity to a GORM model.
func fromDomain(loan *domain.Loan) *GormLoan {
	return &GormLoan{
		ID:            loan.ID,
		UserID:        loan.UserID,
		MRP:           loan.MRP,
		DP:            loan.DP,
		VehicleYear:   loan.VehicleYear,
		PoliceNumber:  loan.PoliceNumber,
		MachineNumber: loan.MachineNumber,
		Status:        loan.Status,
		CreatedAt:     loan.CreatedAt,
		UpdatedAt:     loan.UpdatedAt,
	}
}

// GormLoanRepository implements LoanRepository using GORM.
type GormLoanRepository struct {
	db *gorm.DB
}

// NewGormLoanRepository creates a new GORM-backed loan repository.
func NewGormLoanRepository(db *gorm.DB) *GormLoanRepository {
	return &GormLoanRepository{db: db}
}

// Create inserts a new loan into the database.
func (r *GormLoanRepository) Create(ctx context.Context, loan *domain.Loan) error {
	gormLoan := fromDomain(loan)
	if err := r.db.WithContext(ctx).Create(gormLoan).Error; err != nil {
		return fmt.Errorf("inserting loan: %w", err)
	}
	// Copy back auto-set fields (ID, timestamps)
	loan.ID = gormLoan.ID
	loan.CreatedAt = gormLoan.CreatedAt
	loan.UpdatedAt = gormLoan.UpdatedAt
	return nil
}

// UpdateStatus changes the status of a loan identified by user_id and police_number.
func (r *GormLoanRepository) UpdateStatus(ctx context.Context, userID, policeNumber string, status domain.LoanStatus) error {
	result := r.db.WithContext(ctx).
		Model(&GormLoan{}).
		Where("user_id = ? AND police_number = ?", userID, policeNumber).
		Update("status", string(status))
	if result.Error != nil {
		return fmt.Errorf("updating loan status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("loan not found for user %s and police %s", userID, policeNumber)
	}
	return nil
}

// FindByUserIDAndPoliceNumber retrieves a single loan by its composite key.
func (r *GormLoanRepository) FindByUserIDAndPoliceNumber(ctx context.Context, userID, policeNumber string) (*domain.Loan, error) {
	var gormLoan GormLoan
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND police_number = ?", userID, policeNumber).
		First(&gormLoan).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("querying loan: %w", err)
	}
	return gormLoan.toDomain(), nil
}

// FindByUserID retrieves all loans belonging to a user.
func (r *GormLoanRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Loan, error) {
	var gormLoans []GormLoan
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&gormLoans).Error
	if err != nil {
		return nil, fmt.Errorf("querying loans for user %s: %w", userID, err)
	}

	loans := make([]*domain.Loan, 0, len(gormLoans))
	for i := range gormLoans {
		loans = append(loans, gormLoans[i].toDomain())
	}
	return loans, nil
}
