package repository

import (
	"context"

	"github.com/featureflags/feature-api/internal/models"
)

// FlagRepository defines the storage contract for feature flags.
// Handlers depend on this interface, not on concrete infrastructure.
type FlagRepository interface {
	List(ctx context.Context) ([]models.Flag, error)
	GetByID(ctx context.Context, id string) (*models.Flag, error)
	Create(ctx context.Context, req models.CreateFlagRequest) (*models.Flag, error)
	Update(ctx context.Context, id string, req models.UpdateFlagRequest) (*models.Flag, error)
	Delete(ctx context.Context, id string) error
}
