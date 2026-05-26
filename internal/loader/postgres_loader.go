package loader

import (
	"context"
	"fmt"

	"github.com/wyzauto/translation-loader/internal/domain"
	"github.com/wyzauto/translation-loader/internal/repository"
)

type PostgresLoader struct {
	repo repository.TranslationRepository
}

func NewPostgresLoader(repo repository.TranslationRepository) *PostgresLoader {
	return &PostgresLoader{repo: repo}
}

func (l *PostgresLoader) Load(ctx context.Context, entityType string, entityIDs []string, locales []string) (map[string][]domain.Translation, error) {
	if len(entityIDs) == 0 {
		return map[string][]domain.Translation{}, nil
	}

	rows, err := l.repo.FindByEntities(ctx, entityType, entityIDs, locales)
	if err != nil {
		return nil, fmt.Errorf("load translations for entityType=%s ids=%d: %w", entityType, len(entityIDs), err)
	}

	out := make(map[string][]domain.Translation, len(entityIDs))
	for _, id := range entityIDs {
		out[id] = nil
	}
	for _, tr := range rows {
		out[tr.EntityID] = append(out[tr.EntityID], tr)
	}
	return out, nil
}

func (l *PostgresLoader) Invalidate(entityID string) {}
