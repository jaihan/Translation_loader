package domain

import "time"

type EntityType string

const (
	EntityProduct              EntityType = "product"
	EntityAttribute            EntityType = "attribute"
	EntityProductSpecification EntityType = "product_specification"
)

type Translation struct {
	EntityType string
	EntityID   string
	Locale     string
	FieldName  string
	Value      string
	UpdatedAt  time.Time
}
