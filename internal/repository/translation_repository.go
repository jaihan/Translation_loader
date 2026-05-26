package repository

import (
	"context"

	"github.com/wyzauto/translation-loader/internal/domain"
)

type TranslationRepository interface {
	FindByEntities(ctx context.Context, entityType string, entityIDs []string, locales []string) ([]domain.Translation, error)
}
