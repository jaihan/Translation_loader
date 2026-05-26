package service

import (
	"context"
	"fmt"

	"github.com/wyzauto/translation-loader/internal/domain"
	"github.com/wyzauto/translation-loader/internal/loader"
	"github.com/wyzauto/translation-loader/internal/repository"
)

type ProductDocumentBuilder struct {
	products repository.ProductRepository
	loader   loader.TranslationLoader
	locales  []string
}

func NewProductDocumentBuilder(products repository.ProductRepository, loader loader.TranslationLoader, locales []string) *ProductDocumentBuilder {
	return &ProductDocumentBuilder{
		products: products,
		loader:   loader,
		locales:  locales,
	}
}

func (b *ProductDocumentBuilder) Build(ctx context.Context, productID string) (domain.ProductDocument, error) {
	product, err := b.products.FindProduct(ctx, productID)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("find product %s: %w", productID, err)
	}

	specs, err := b.products.FindSpecifications(ctx, productID)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("find specifications %s: %w", productID, err)
	}

	productTrans, err := b.loader.Load(ctx, string(domain.EntityProduct), []string{product.ID}, b.locales)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("load product translations %s: %w", product.ID, err)
	}

	attrIDs := make([]string, 0, len(specs))
	specIDs := make([]string, 0, len(specs))
	for _, s := range specs {
		attrIDs = append(attrIDs, s.AttributeID)
		specIDs = append(specIDs, s.ID)
	}

	attrTrans := map[string][]domain.Translation{}
	if len(attrIDs) > 0 {
		attrTrans, err = b.loader.Load(ctx, string(domain.EntityAttribute), attrIDs, b.locales)
		if err != nil {
			return domain.ProductDocument{}, fmt.Errorf("load attribute translations for product %s: %w", product.ID, err)
		}
	}

	specTrans := map[string][]domain.Translation{}
	if len(specIDs) > 0 {
		specTrans, err = b.loader.Load(ctx, string(domain.EntityProductSpecification), specIDs, b.locales)
		if err != nil {
			return domain.ProductDocument{}, fmt.Errorf("load specification translations for product %s: %w", product.ID, err)
		}
	}

	doc := domain.ProductDocument{
		UUID:       product.ID,
		SKU:        product.SKU,
		PartNumber: product.PartNumber,
		Brand: domain.BrandDocument{
			Code:  product.Brand,
			Label: map[string]string{},
		},
		Attributes: map[string]string{},
		Values:     map[string]domain.AttributeValueDocument{},
	}

	productNames := fieldMap(productTrans[product.ID], "productname")
	for _, locale := range b.locales {
		doc.ProductName = append(doc.ProductName, domain.LocalizedText{
			Locale: locale,
			Data:   fallback(productNames, locale),
		})
	}

	for _, locale := range b.locales {
		doc.Brand.Label[locale] = product.Brand
	}

	for _, spec := range specs {
		code := spec.Attribute.Code
		doc.Attributes[code] = spec.Value

		attrLabels := fieldMap(attrTrans[spec.AttributeID], "label")
		valueLabels := fieldMap(specTrans[spec.ID], "value_label")

		labels := make(map[string]string, len(b.locales))
		for _, locale := range b.locales {
			label := fallback(valueLabels, locale)
			if label == "" {
				label = fallback(attrLabels, locale)
			}
			labels[locale] = label
		}

		doc.Values[code] = domain.AttributeValueDocument{
			Code:  spec.Value,
			Label: labels,
		}
	}

	return doc, nil
}

func fieldMap(rows []domain.Translation, fieldName string) map[string]string {
	out := make(map[string]string)
	for _, r := range rows {
		if r.FieldName == fieldName {
			out[r.Locale] = r.Value
		}
	}
	return out
}

func fallback(values map[string]string, locale string) string {
	if values == nil {
		return ""
	}
	if v, ok := values[locale]; ok {
		return v
	}
	if v, ok := values["en"]; ok {
		return v
	}
	return ""
}
