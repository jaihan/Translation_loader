package unit

import (
	"context"
	"testing"

	"github.com/wyzauto/translation-loader/internal/domain"
	"github.com/wyzauto/translation-loader/internal/service"
)

type mockProductRepository struct{}

func (m mockProductRepository) FindProduct(ctx context.Context, productID string) (domain.Product, error) {
	return domain.Product{
		ID:         productID,
		SKU:        "BRK-PAD-HILUX-FRT-04465-0K390",
		PartNumber: "04465-0K390",
		Brand:      "bosch",
		CategoryID: "brake-system",
	}, nil
}

func (m mockProductRepository) FindSpecifications(ctx context.Context, productID string) ([]domain.ProductSpecification, error) {
	return []domain.ProductSpecification{
		{
			ID:          "spec-position",
			ProductID:   productID,
			AttributeID: "attr-position",
			Value:       "front",
			Attribute: domain.Attribute{
				ID:   "attr-position",
				Code: "position",
			},
		},
		{
			ID:          "spec-material",
			ProductID:   productID,
			AttributeID: "attr-material",
			Value:       "ceramic",
			Attribute: domain.Attribute{
				ID:   "attr-material",
				Code: "brake_pad_material",
			},
		},
	}, nil
}

type mockTranslationLoader struct{}

func (m mockTranslationLoader) Load(
	ctx context.Context,
	entityType string,
	entityIDs []string,
	locales []string,
) (map[string][]domain.Translation, error) {
	out := map[string][]domain.Translation{}

	switch entityType {
	case "product":
		out["product-1"] = []domain.Translation{
			{
				EntityType: "product",
				EntityID:   "product-1",
				Locale:     "en",
				FieldName:  "productname",
				Value:      "Bosch Front Brake Pad Set for Toyota Hilux Revo",
			},
			{
				EntityType: "product",
				EntityID:   "product-1",
				Locale:     "th",
				FieldName:  "productname",
				Value:      "ผ้าเบรกหน้า Bosch สำหรับ Toyota Hilux Revo",
			},
		}

	case "attribute":
		out["attr-position"] = []domain.Translation{
			{
				EntityType: "attribute",
				EntityID:   "attr-position",
				Locale:     "en",
				FieldName:  "label",
				Value:      "Position",
			},
			{
				EntityType: "attribute",
				EntityID:   "attr-position",
				Locale:     "th",
				FieldName:  "label",
				Value:      "ตำแหน่งติดตั้ง",
			},
		}

		out["attr-material"] = []domain.Translation{
			{
				EntityType: "attribute",
				EntityID:   "attr-material",
				Locale:     "en",
				FieldName:  "label",
				Value:      "Brake Pad Material",
			},
			{
				EntityType: "attribute",
				EntityID:   "attr-material",
				Locale:     "th",
				FieldName:  "label",
				Value:      "วัสดุผ้าเบรก",
			},
		}

	case "product_specification":
		out["spec-position"] = []domain.Translation{
			{
				EntityType: "product_specification",
				EntityID:   "spec-position",
				Locale:     "en",
				FieldName:  "value_label",
				Value:      "Front",
			},
			{
				EntityType: "product_specification",
				EntityID:   "spec-position",
				Locale:     "th",
				FieldName:  "value_label",
				Value:      "ด้านหน้า",
			},
		}

		out["spec-material"] = []domain.Translation{
			{
				EntityType: "product_specification",
				EntityID:   "spec-material",
				Locale:     "en",
				FieldName:  "value_label",
				Value:      "Ceramic",
			},
			{
				EntityType: "product_specification",
				EntityID:   "spec-material",
				Locale:     "th",
				FieldName:  "value_label",
				Value:      "เซรามิก",
			},
		}
	}

	return out, nil
}

func (m mockTranslationLoader) Invalidate(entityID string) {}

func TestProductDocumentBuilderBuildsLocalizedDocument(t *testing.T) {
	ctx := context.Background()

	builder := service.NewProductDocumentBuilder(
		mockProductRepository{},
		mockTranslationLoader{},
		[]string{"en", "th"},
	)

	doc, err := builder.Build(ctx, "product-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if doc.UUID != "product-1" {
		t.Fatalf("expected product-1, got %s", doc.UUID)
	}

	if doc.SKU != "BRK-PAD-HILUX-FRT-04465-0K390" {
		t.Fatalf("unexpected SKU: %s", doc.SKU)
	}

	if doc.PartNumber != "04465-0K390" {
		t.Fatalf("unexpected part number: %s", doc.PartNumber)
	}

	if doc.Brand.Code != "bosch" {
		t.Fatalf("expected brand bosch, got %s", doc.Brand.Code)
	}

	if len(doc.ProductName) != 2 {
		t.Fatalf("expected 2 productname values, got %d", len(doc.ProductName))
	}

	if doc.ProductName[0].Locale != "en" {
		t.Fatalf("expected first locale en, got %s", doc.ProductName[0].Locale)
	}

	if doc.ProductName[0].Data != "Bosch Front Brake Pad Set for Toyota Hilux Revo" {
		t.Fatalf("unexpected en product name: %s", doc.ProductName[0].Data)
	}

	if doc.ProductName[1].Locale != "th" {
		t.Fatalf("expected second locale th, got %s", doc.ProductName[1].Locale)
	}

	if doc.ProductName[1].Data != "ผ้าเบรกหน้า Bosch สำหรับ Toyota Hilux Revo" {
		t.Fatalf("unexpected th product name: %s", doc.ProductName[1].Data)
	}

	if doc.Attributes["position"] != "front" {
		t.Fatalf("expected position front, got %s", doc.Attributes["position"])
	}

	if doc.Attributes["brake_pad_material"] != "ceramic" {
		t.Fatalf("expected brake_pad_material ceramic, got %s", doc.Attributes["brake_pad_material"])
	}

	if doc.Values["position"].Code != "front" {
		t.Fatalf("expected position code front, got %s", doc.Values["position"].Code)
	}

	if doc.Values["position"].Label["en"] != "Front" {
		t.Fatalf("expected position en label Front, got %s", doc.Values["position"].Label["en"])
	}

	if doc.Values["position"].Label["th"] != "ด้านหน้า" {
		t.Fatalf("expected position th label ด้านหน้า, got %s", doc.Values["position"].Label["th"])
	}

	if doc.Values["brake_pad_material"].Code != "ceramic" {
		t.Fatalf("expected material code ceramic, got %s", doc.Values["brake_pad_material"].Code)
	}

	if doc.Values["brake_pad_material"].Label["en"] != "Ceramic" {
		t.Fatalf("expected material en label Ceramic, got %s", doc.Values["brake_pad_material"].Label["en"])
	}

	if doc.Values["brake_pad_material"].Label["th"] != "เซรามิก" {
		t.Fatalf("expected material th label เซรามิก, got %s", doc.Values["brake_pad_material"].Label["th"])
	}
}
