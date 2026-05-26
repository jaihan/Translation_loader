package integration

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wyzauto/translation-loader/internal/cache"
	"github.com/wyzauto/translation-loader/internal/loader"
	"github.com/wyzauto/translation-loader/internal/repository"
	"github.com/wyzauto/translation-loader/internal/service"
)

func TestProductDocumentBuilderIntegration(t *testing.T) {
	ctx := context.Background()

	db := setupTestDatabase(t, ctx)
	defer db.Close()

	productID := "11111111-1111-1111-1111-111111111111"

	seedTestData(t, ctx, db, productID)

	productRepo := repository.NewPostgresProductRepository(db)
	translationRepo := repository.NewPostgresTranslationRepository(db)

	baseLoader := loader.NewPostgresLoader(translationRepo)
	ttlCache := cache.NewTTLCache(5 * time.Second)
	cachedLoader := loader.NewCachedLoader(baseLoader, ttlCache)

	builder := service.NewProductDocumentBuilder(
		productRepo,
		cachedLoader,
		[]string{"en", "th"},
	)

	doc, err := builder.Build(ctx, productID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if doc.UUID != productID {
		t.Fatalf("expected UUID %s, got %s", productID, doc.UUID)
	}

	if doc.SKU != "BP-OIL-5W30-1L" {
		t.Fatalf("expected SKU BP-OIL-5W30-1L, got %s", doc.SKU)
	}

	if doc.PartNumber != "5W30-1L" {
		t.Fatalf("expected part number 5W30-1L, got %s", doc.PartNumber)
	}

	if doc.Brand.Code != "bosch" {
		t.Fatalf("expected brand bosch, got %s", doc.Brand.Code)
	}

	if len(doc.ProductName) == 0 {
		t.Fatal("expected productname translations")
	}

	if doc.Attributes["oil_grade"] != "5w30" {
		t.Fatalf("expected oil_grade 5w30, got %s", doc.Attributes["oil_grade"])
	}
}

func setupTestDatabase(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	dsn := "postgres://wyzauto:wyzauto@localhost:5432/wyzauto?sslmode=disable"

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}

	statements := []string{
		`DELETE FROM translation`,
		`DELETE FROM product_specification`,
		`DELETE FROM attribute`,
		`DELETE FROM product`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(ctx, stmt); err != nil {
			t.Fatalf("cleanup failed: %v", err)
		}
	}

	return db
}

func seedTestData(t *testing.T, ctx context.Context, db *pgxpool.Pool, productID string) {
	t.Helper()

	categoryID := "22222222-2222-2222-2222-222222222222"
	attributeID := "33333333-3333-3333-3333-333333333333"
	specID := "44444444-4444-4444-4444-444444444444"

	queries := []struct {
		sql  string
		args []any
	}{
		{
			sql: `INSERT INTO product (id, sku, part_number, brand, category_id)
			      VALUES ($1, $2, $3, $4, $5)`,
			args: []any{productID, "BP-OIL-5W30-1L", "5W30-1L", "bosch", categoryID},
		},
		{
			sql: `INSERT INTO attribute (id, code, metric_unit)
			      VALUES ($1, $2, $3)`,
			args: []any{attributeID, "oil_grade", nil},
		},
		{
			sql: `INSERT INTO product_specification (id, product_id, attribute_id, value)
			      VALUES ($1, $2, $3, $4)`,
			args: []any{specID, productID, attributeID, "5w30"},
		},
		{
			sql: `INSERT INTO translation (id, entity_type, entity_id, locale, field_name, field_value, updated_at)
			      VALUES ($1, 'product', $2, 'en', 'productname', '5W-30 Engine Oil 1L', NOW())`,
			args: []any{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1", productID},
		},
		{
			sql: `INSERT INTO translation (id, entity_type, entity_id, locale, field_name, field_value, updated_at)
			      VALUES ($1, 'product', $2, 'th', 'productname', 'น้ำมันเครื่อง 5W-30 1 ลิตร', NOW())`,
			args: []any{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2", productID},
		},
		{
			sql: `INSERT INTO translation (id, entity_type, entity_id, locale, field_name, field_value, updated_at)
			      VALUES ($1, 'attribute', $2, 'en', 'label', 'Oil Grade', NOW())`,
			args: []any{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3", attributeID},
		},
		{
			sql: `INSERT INTO translation (id, entity_type, entity_id, locale, field_name, field_value, updated_at)
			      VALUES ($1, 'product_specification', $2, 'en', 'value_label', '5W-30', NOW())`,
			args: []any{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4", specID},
		},
	}

	for _, q := range queries {
		if _, err := db.Exec(ctx, q.sql, q.args...); err != nil {
			t.Fatalf("seed failed: %v", err)
		}
	}
}