package usecase

import (
	"context"
	"fmt"

	branchDomain "juansecalvinio/tepidolacuenta/internal/branch/domain"
	branchRepo "juansecalvinio/tepidolacuenta/internal/branch/repository"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	restaurantDomain "juansecalvinio/tepidolacuenta/internal/restaurant/domain"
	restaurantRepo "juansecalvinio/tepidolacuenta/internal/restaurant/repository"
	"juansecalvinio/tepidolacuenta/internal/setup/domain"
	tableDomain "juansecalvinio/tepidolacuenta/internal/table/domain"
	tableRepo "juansecalvinio/tepidolacuenta/internal/table/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UseCase defines the interface for setup operations
type UseCase interface {
	SetupRestaurant(ctx context.Context, userID primitive.ObjectID, input domain.SetupRestaurantInput) (*domain.SetupRestaurantOutput, error)
}

type setupUseCase struct {
	restaurantRepo restaurantRepo.Repository
	branchRepo     branchRepo.Repository
	tableRepo      tableRepo.Repository
	qrService      *pkg.QRService
}

// NewSetupUseCase creates a new setup use case
func NewSetupUseCase(
	restaurantRepo restaurantRepo.Repository,
	branchRepo branchRepo.Repository,
	tableRepo tableRepo.Repository,
	qrService *pkg.QRService,
) UseCase {
	return &setupUseCase{
		restaurantRepo: restaurantRepo,
		branchRepo:     branchRepo,
		tableRepo:      tableRepo,
		qrService:      qrService,
	}
}

// SetupRestaurant creates a restaurant, its first branch, and tables in a single operation
func (uc *setupUseCase) SetupRestaurant(ctx context.Context, userID primitive.ObjectID, input domain.SetupRestaurantInput) (*domain.SetupRestaurantOutput, error) {
	// Step 1: Create restaurant
	restaurant := restaurantDomain.NewRestaurant(userID, input.Name, input.CUIT)
	if err := uc.restaurantRepo.Create(ctx, restaurant); err != nil {
		return nil, fmt.Errorf("failed to create restaurant: %w", err)
	}

	// Step 2: Create branch linked to the new restaurant
	branch := branchDomain.NewBranch(restaurant.ID, input.Address, "")
	if err := uc.branchRepo.Create(ctx, branch); err != nil {
		_ = uc.restaurantRepo.Delete(ctx, restaurant.ID)
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Step 3: Create tables linked to the new branch
	tables := make([]*tableDomain.Table, 0, input.TableCount)
	for i := 1; i <= input.TableCount; i++ {
		newTable := tableDomain.NewTable(branch.ID, i, "")

		if err := uc.tableRepo.Create(ctx, newTable); err != nil {
			uc.cleanupTables(ctx, tables)
			_ = uc.branchRepo.Delete(ctx, branch.ID)
			_ = uc.restaurantRepo.Delete(ctx, restaurant.ID)
			return nil, fmt.Errorf("failed to create table %d: %w", i, err)
		}

		// Generate QR code with the actual table ID
		qrCode := uc.qrService.GenerateTableQRCode(restaurant.ID, branch.ID, newTable.ID, i)
		newTable.QRCode = qrCode

		if err := uc.tableRepo.Update(ctx, newTable); err != nil {
			uc.cleanupTables(ctx, tables)
			_ = uc.tableRepo.Delete(ctx, newTable.ID)
			_ = uc.branchRepo.Delete(ctx, branch.ID)
			_ = uc.restaurantRepo.Delete(ctx, restaurant.ID)
			return nil, fmt.Errorf("failed to update QR code for table %d: %w", i, err)
		}

		tables = append(tables, newTable)
	}

	return &domain.SetupRestaurantOutput{
		Restaurant: restaurant,
		Branch:     branch,
		Tables:     tables,
	}, nil
}

// cleanupTables deletes all tables that were created during a failed setup
func (uc *setupUseCase) cleanupTables(ctx context.Context, tables []*tableDomain.Table) {
	for _, t := range tables {
		_ = uc.tableRepo.Delete(ctx, t.ID)
	}
}
