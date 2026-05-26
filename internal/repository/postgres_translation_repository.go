package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wyzauto/translation-loader/internal/domain"
)

type PostgresTranslationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTranslationRepository(pool *pgxpool.Pool) *PostgresTranslationRepository {
	return &PostgresTranslationRepository{pool: pool}
}

func (r *PostgresTranslationRepository) FindByEntities(ctx context.Context, entityType string, entityIDs []string, locales []string) ([]domain.Translation, error) {
	if len(entityIDs) == 0 || len(locales) == 0 {
		return nil, nil
	}

	const q = `
SELECT entity_type, entity_id, locale, field_name, field_value, updated_at
FROM translation
WHERE entity_type = $1
  AND entity_id = ANY($2)
  AND locale = ANY($3)
ORDER BY entity_id, locale, field_name`

	rows, err := r.pool.Query(ctx, q, entityType, entityIDs, locales)
	if err != nil {
		return nil, fmt.Errorf("query translations for %s: %w", entityType, err)
	}
	defer rows.Close()

	var result []domain.Translation
	for rows.Next() {
		var t domain.Translation
		if err := rows.Scan(&t.EntityType, &t.EntityID, &t.Locale, &t.FieldName, &t.Value, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan translation: %w", err)
		}
		result = append(result, t)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate translations: %w", rows.Err())
	}
	return result, nil
}
