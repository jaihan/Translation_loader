package loader

import (
	"context"

	"github.com/wyzauto/translation-loader/internal/domain"
)

type TranslationLoader interface {
	Load(ctx context.Context, entityType string, entityIDs []string, locales []string) (map[string][]domain.Translation, error)
	Invalidate(entityID string)
}
