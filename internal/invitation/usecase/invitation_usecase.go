package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"juansecalvinio/tepidolacuenta/internal/invitation/domain"
	invitationRepo "juansecalvinio/tepidolacuenta/internal/invitation/repository"
	"juansecalvinio/tepidolacuenta/internal/pkg"
	restaurantRepo "juansecalvinio/tepidolacuenta/internal/restaurant/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UseCase interface {
	Generate(ctx context.Context, ownerUserID, restaurantID primitive.ObjectID) (*domain.GenerateInvitationResponse, error)
	Redeem(ctx context.Context, code string) (*domain.Invitation, error)
}

type invitationUseCase struct {
	repo           invitationRepo.Repository
	restaurantRepo restaurantRepo.Repository
}

func NewInvitationUseCase(repo invitationRepo.Repository, restaurantRepo restaurantRepo.Repository) UseCase {
	return &invitationUseCase{repo: repo, restaurantRepo: restaurantRepo}
}

func (uc *invitationUseCase) Generate(ctx context.Context, ownerUserID, restaurantID primitive.ObjectID) (*domain.GenerateInvitationResponse, error) {
	restaurant, err := uc.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	if restaurant.UserID != ownerUserID {
		return nil, pkg.ErrUnauthorized
	}

	code, err := generateCode()
	if err != nil {
		return nil, pkg.ErrInternalServer
	}

	inv := domain.NewInvitation(restaurantID, code)
	if err := uc.repo.Create(ctx, inv); err != nil {
		return nil, err
	}

	return &domain.GenerateInvitationResponse{Code: inv.Code, ExpiresAt: inv.ExpiresAt}, nil
}

func (uc *invitationUseCase) Redeem(ctx context.Context, code string) (*domain.Invitation, error) {
	inv, err := uc.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, pkg.ErrInvalidToken
	}
	if inv.Used {
		return nil, pkg.ErrInvalidToken
	}
	if time.Now().After(inv.ExpiresAt) {
		return nil, pkg.ErrResetTokenExpired
	}
	if err := uc.repo.MarkUsed(ctx, code); err != nil {
		return nil, err
	}
	return inv, nil
}

func generateCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
