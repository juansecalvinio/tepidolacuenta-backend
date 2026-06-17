package repository

import (
	"context"

	"juansecalvinio/tepidolacuenta/internal/invitation/domain"
)

type Repository interface {
	Create(ctx context.Context, inv *domain.Invitation) error
	FindByCode(ctx context.Context, code string) (*domain.Invitation, error)
	MarkUsed(ctx context.Context, code string) error
}
