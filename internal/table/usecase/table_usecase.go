package usecase

import (
	"context"
	"errors"
	"fmt"

	branchRepo "juansecalvinio/tepidolacuenta/internal/branch/repository"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	restaurantRepo "juansecalvinio/tepidolacuenta/internal/restaurant/repository"
	"juansecalvinio/tepidolacuenta/internal/table/domain"
	"juansecalvinio/tepidolacuenta/internal/table/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UseCase defines the interface for table use cases
type UseCase interface {
	Create(ctx context.Context, userID primitive.ObjectID, input domain.CreateTableInput) (*domain.Table, error)
	BulkCreate(ctx context.Context, userID primitive.ObjectID, input domain.BulkCreateTablesInput) ([]*domain.Table, error)
	GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*domain.Table, error)
	GetByBranchID(ctx context.Context, branchID primitive.ObjectID, userID primitive.ObjectID) ([]*domain.Table, error)
	Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateTableInput) (*domain.Table, error)
	Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
}

type tableUseCase struct {
	repo           repository.Repository
	branchRepo     branchRepo.Repository
	restaurantRepo restaurantRepo.Repository
	qrService      *pkg.QRService
}

// NewTableUseCase creates a new table use case
func NewTableUseCase(repo repository.Repository, branchRepo branchRepo.Repository, restaurantRepo restaurantRepo.Repository, qrService *pkg.QRService) UseCase {
	return &tableUseCase{
		repo:           repo,
		branchRepo:     branchRepo,
		restaurantRepo: restaurantRepo,
		qrService:      qrService,
	}
}

// verifyBranchOwnership verifies that a branch exists and the user owns it
func (uc *tableUseCase) verifyBranchOwnership(ctx context.Context, branchID, userID primitive.ObjectID) (*primitive.ObjectID, error) {
	branch, err := uc.branchRepo.FindByID(ctx, branchID)
	if err != nil {
		return nil, err
	}

	restaurant, err := uc.restaurantRepo.FindByID(ctx, branch.RestaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	return &branch.RestaurantID, nil
}

// Create creates a new table
func (uc *tableUseCase) Create(ctx context.Context, userID primitive.ObjectID, input domain.CreateTableInput) (*domain.Table, error) {
	// Validate input
	if !pkg.IsValidTableNumber(input.Number) {
		return nil, errors.New("table number must be greater than 0")
	}

	// Parse branch ID
	branchID, err := primitive.ObjectIDFromHex(input.BranchID)
	if err != nil {
		return nil, errors.New("invalid branch ID")
	}

	// Verify branch exists and user owns it
	restaurantID, err := uc.verifyBranchOwnership(ctx, branchID, userID)
	if err != nil {
		return nil, err
	}

	// Check if table number already exists for this branch
	existingTable, err := uc.repo.FindByBranchAndNumber(ctx, branchID, input.Number)
	if err != nil {
		return nil, err
	}

	if existingTable != nil {
		return nil, fmt.Errorf("table number %d already exists for this branch", input.Number)
	}

	// Create table with temporary ID for QR generation
	tempTable := domain.NewTable(branchID, input.Number, "")

	// Save to get the actual ID
	if err := uc.repo.Create(ctx, tempTable); err != nil {
		return nil, err
	}

	// Generate QR code with the actual table ID
	qrCode := uc.qrService.GenerateTableQRCode(*restaurantID, branchID, tempTable.ID, input.Number)
	tempTable.QRCode = qrCode

	// Update with QR code
	if err := uc.repo.Update(ctx, tempTable); err != nil {
		return nil, err
	}

	return tempTable, nil
}

// BulkCreate creates multiple tables for a branch
func (uc *tableUseCase) BulkCreate(ctx context.Context, userID primitive.ObjectID, input domain.BulkCreateTablesInput) ([]*domain.Table, error) {
	// Parse branch ID
	branchID, err := primitive.ObjectIDFromHex(input.BranchID)
	if err != nil {
		return nil, errors.New("invalid branch ID")
	}

	// Verify branch exists and user owns it
	restaurantID, err := uc.verifyBranchOwnership(ctx, branchID, userID)
	if err != nil {
		return nil, err
	}

	// Get existing tables to determine starting number
	existingTables, err := uc.repo.FindByBranchID(ctx, branchID)
	if err != nil {
		return nil, err
	}

	// Find the highest table number
	maxTableNumber := 0
	for _, table := range existingTables {
		if table.Number > maxTableNumber {
			maxTableNumber = table.Number
		}
	}

	// Create tables
	createdTables := make([]*domain.Table, 0, input.Count)
	for i := 1; i <= input.Count; i++ {
		tableNumber := maxTableNumber + i

		// Create table
		newTable := domain.NewTable(branchID, tableNumber, "")

		// Save to get the actual ID
		if err := uc.repo.Create(ctx, newTable); err != nil {
			return nil, fmt.Errorf("failed to create table %d: %w", tableNumber, err)
		}

		// Generate QR code with the actual table ID
		qrCode := uc.qrService.GenerateTableQRCode(*restaurantID, branchID, newTable.ID, tableNumber)
		newTable.QRCode = qrCode

		// Update with QR code
		if err := uc.repo.Update(ctx, newTable); err != nil {
			return nil, fmt.Errorf("failed to update QR code for table %d: %w", tableNumber, err)
		}

		createdTables = append(createdTables, newTable)
	}

	return createdTables, nil
}

// GetByID retrieves a table by ID
func (uc *tableUseCase) GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*domain.Table, error) {
	table, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify user owns the branch
	_, err = uc.verifyBranchOwnership(ctx, table.BranchID, userID)
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByBranchID retrieves all tables for a branch
func (uc *tableUseCase) GetByBranchID(ctx context.Context, branchID primitive.ObjectID, userID primitive.ObjectID) ([]*domain.Table, error) {
	// Verify user owns the branch
	_, err := uc.verifyBranchOwnership(ctx, branchID, userID)
	if err != nil {
		return nil, err
	}

	return uc.repo.FindByBranchID(ctx, branchID)
}

// Update updates a table
func (uc *tableUseCase) Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateTableInput) (*domain.Table, error) {
	// Find table
	table, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify user owns the branch
	restaurantID, err := uc.verifyBranchOwnership(ctx, table.BranchID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	needsQRUpdate := false

	if input.Number != 0 {
		if !pkg.IsValidTableNumber(input.Number) {
			return nil, errors.New("table number must be greater than 0")
		}

		// Check if new number already exists (excluding current table)
		existingTable, err := uc.repo.FindByBranchAndNumber(ctx, table.BranchID, input.Number)
		if err != nil {
			return nil, err
		}

		if existingTable != nil && existingTable.ID != table.ID {
			return nil, fmt.Errorf("table number %d already exists for this branch", input.Number)
		}

		table.Number = input.Number
		needsQRUpdate = true
	}

	if input.IsActive != nil {
		table.IsActive = *input.IsActive
	}

	// Regenerate QR code if number changed
	if needsQRUpdate {
		qrCode := uc.qrService.GenerateTableQRCode(*restaurantID, table.BranchID, table.ID, table.Number)
		table.QRCode = qrCode
	}

	// Save changes
	if err := uc.repo.Update(ctx, table); err != nil {
		return nil, err
	}

	return table, nil
}

// Delete deletes a table
func (uc *tableUseCase) Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	// Find table
	table, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify user owns the branch
	_, err = uc.verifyBranchOwnership(ctx, table.BranchID, userID)
	if err != nil {
		return err
	}

	return uc.repo.Delete(ctx, id)
}
