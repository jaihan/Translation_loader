package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wyzauto/translation-loader/internal/domain"
)

type ProductRepository interface {
	FindProduct(ctx context.Context, productID string) (domain.Product, error)
	FindSpecifications(ctx context.Context, productID string) ([]domain.ProductSpecification, error)
}

type PostgresProductRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProductRepository(db *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) FindProduct(ctx context.Context, productID string) (domain.Product, error) {
	var p domain.Product

	err := r.db.QueryRow(ctx, `
		SELECT id, sku, part_number, brand, category_id
		FROM product
		WHERE id = $1
	`, productID).Scan(
		&p.ID,
		&p.SKU,
		&p.PartNumber,
		&p.Brand,
		&p.CategoryID,
	)

	if err != nil {
		return domain.Product{}, fmt.Errorf("query product: %w", err)
	}

	return p, nil
}

func (r *PostgresProductRepository) FindSpecifications(ctx context.Context, productID string) ([]domain.ProductSpecification, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
			ps.id,
			ps.product_id,
			ps.attribute_id,
			ps.value,
			a.id,
			a.code,
			a.metric_unit
		FROM product_specification ps
		JOIN attribute a ON a.id = ps.attribute_id
		WHERE ps.product_id = $1
	`, productID)
	if err != nil {
		return nil, fmt.Errorf("query product specifications: %w", err)
	}
	defer rows.Close()

	var specs []domain.ProductSpecification

	for rows.Next() {
		var s domain.ProductSpecification

		err := rows.Scan(
			&s.ID,
			&s.ProductID,
			&s.AttributeID,
			&s.Value,
			&s.Attribute.ID,
			&s.Attribute.Code,
			&s.Attribute.MetricUnit,
		)
		if err != nil {
			return nil, fmt.Errorf("scan product specification: %w", err)
		}

		specs = append(specs, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product specifications: %w", err)
	}

	return specs, nil
}