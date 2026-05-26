package domain

type Product struct {
	ID         string
	SKU        string
	PartNumber string
	Brand      string
	CategoryID string
}

type Attribute struct {
	ID         string
	Code       string
	MetricUnit *string
}

type ProductSpecification struct {
	ID          string
	ProductID   string
	AttributeID string
	Value       string
	Attribute   Attribute
}
