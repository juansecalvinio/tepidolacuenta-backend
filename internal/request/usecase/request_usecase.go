package usecase

import (
	"context"
	"errors"

	"juansecalvinio/tepidolacuenta/internal/pkg"
	"juansecalvinio/tepidolacuenta/internal/request/domain"
	"juansecalvinio/tepidolacuenta/internal/request/repository"
	restaurantRepo "juansecalvinio/tepidolacuenta/internal/restaurant/repository"
	tableRepo "juansecalvinio/tepidolacuenta/internal/table/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UseCase defines the interface for request use cases
type UseCase interface {
	Create(ctx context.Context, input domain.CreateRequestInput) (*domain.Request, error)
	GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*domain.Request, error)
	GetByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID, userID primitive.ObjectID) ([]*domain.Request, error)
	GetPendingByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID, userID primitive.ObjectID) ([]*domain.Request, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateRequestStatusInput) (*domain.Request, error)
	Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
}

type requestUseCase struct {
	repo           repository.Repository
	restaurantRepo restaurantRepo.Repository
	tableRepo      tableRepo.Repository
	qrService      *pkg.QRService
	notifyFunc     func(restaurantID primitive.ObjectID, request *domain.Request)
}

// NewRequestUseCase creates a new request use case
func NewRequestUseCase(
	repo repository.Repository,
	restaurantRepo restaurantRepo.Repository,
	tableRepo tableRepo.Repository,
	qrService *pkg.QRService,
	notifyFunc func(restaurantID primitive.ObjectID, request *domain.Request),
) UseCase {
	return &requestUseCase{
		repo:           repo,
		restaurantRepo: restaurantRepo,
		tableRepo:      tableRepo,
		qrService:      qrService,
		notifyFunc:     notifyFunc,
	}
}

// Create creates a new request from a QR code scan
func (uc *requestUseCase) Create(ctx context.Context, input domain.CreateRequestInput) (*domain.Request, error) {
	// Parse IDs
	restaurantID, err := primitive.ObjectIDFromHex(input.RestaurantID)
	if err != nil {
		return nil, errors.New("invalid restaurant ID")
	}

	tableID, err := primitive.ObjectIDFromHex(input.TableID)
	if err != nil {
		return nil, errors.New("invalid table ID")
	}

	// Validate QR code
	if !uc.qrService.ValidateTableQRCode(restaurantID, tableID, input.TableNumber, input.Hash) {
		return nil, errors.New("invalid QR code")
	}

	// Verify restaurant exists
	restaurant, err := uc.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	// Verify table exists and belongs to restaurant
	table, err := uc.tableRepo.FindByID(ctx, tableID)
	if err != nil {
		return nil, err
	}

	if table.RestaurantID != restaurantID {
		return nil, errors.New("table does not belong to restaurant")
	}

	if !table.IsActive {
		return nil, errors.New("table is not active")
	}

	// Create request
	request := domain.NewRequest(restaurantID, tableID, input.TableNumber)

	if err := uc.repo.Create(ctx, request); err != nil {
		return nil, err
	}

	// Notify restaurant via WebSocket
	if uc.notifyFunc != nil {
		uc.notifyFunc(restaurant.ID, request)
	}

	return request, nil
}

// GetByID retrieves a request by ID
func (uc *requestUseCase) GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*domain.Request, error) {
	request, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify user owns the restaurant
	restaurant, err := uc.restaurantRepo.FindByID(ctx, request.RestaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	return request, nil
}

// GetByRestaurantID retrieves all requests for a restaurant
func (uc *requestUseCase) GetByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID, userID primitive.ObjectID) ([]*domain.Request, error) {
	// Verify user owns the restaurant
	restaurant, err := uc.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	return uc.repo.FindByRestaurantID(ctx, restaurantID)
}

// GetPendingByRestaurantID retrieves all pending requests for a restaurant
func (uc *requestUseCase) GetPendingByRestaurantID(ctx context.Context, restaurantID primitive.ObjectID, userID primitive.ObjectID) ([]*domain.Request, error) {
	// Verify user owns the restaurant
	restaurant, err := uc.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	return uc.repo.FindPendingByRestaurantID(ctx, restaurantID)
}

// UpdateStatus updates a request's status
func (uc *requestUseCase) UpdateStatus(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, input domain.UpdateRequestStatusInput) (*domain.Request, error) {
	// Find request
	request, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify user owns the restaurant
	restaurant, err := uc.restaurantRepo.FindByID(ctx, request.RestaurantID)
	if err != nil {
		return nil, err
	}

	if restaurant.UserID != userID {
		return nil, pkg.ErrUnauthorized
	}

	// Update status
	request.UpdateStatus(domain.RequestStatus(input.Status))

	if err := uc.repo.Update(ctx, request); err != nil {
		return nil, err
	}

	return request, nil
}

// Delete deletes a request
func (uc *requestUseCase) Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	// Find request
	request, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify user owns the restaurant
	restaurant, err := uc.restaurantRepo.FindByID(ctx, request.RestaurantID)
	if err != nil {
		return err
	}

	if restaurant.UserID != userID {
		return pkg.ErrUnauthorized
	}

	return uc.repo.Delete(ctx, id)
}
