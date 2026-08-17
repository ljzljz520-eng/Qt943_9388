package stationery

import "errors"

type CategoryAxis string

const (
	CategoryAxisGrade   CategoryAxis = "grade"
	CategoryAxisPurpose CategoryAxis = "purpose"
)

type ProductKind string

const (
	ProductKindNotebook     ProductKind = "notebook"
	ProductKindDrawingPaper ProductKind = "drawing_paper"
	ProductKindMarker       ProductKind = "marker"
	ProductKindFolder       ProductKind = "folder"
)

type Category struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Axis     CategoryAxis `json:"axis"`
	ParentID string       `json:"parentId,omitempty"`
}

type ProductImage struct {
	ID        string `json:"id"`
	ProductID string `json:"productId"`
	URI       string `json:"uri"`
	Alt       string `json:"alt"`
	Position  int    `json:"position"`
}

type Product struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	Kind              ProductKind    `json:"kind"`
	GradeCategoryID   string         `json:"gradeCategoryId"`
	PurposeCategoryID string         `json:"purposeCategoryId"`
	PriceCents        int64          `json:"priceCents"`
	Stock             int            `json:"stock"`
	Images            []ProductImage `json:"images"`
}

type ImageInput struct {
	URI string
	Alt string
}

type AddProductInput struct {
	Name              string
	Kind              ProductKind
	GradeCategoryID   string
	PurposeCategoryID string
	PriceCents        int64
	Stock             int
	Images            []ImageInput
}

type ChangeField string

const (
	ChangeFieldPrice ChangeField = "price_cents"
	ChangeFieldStock ChangeField = "stock"
)

type ChangeRecord struct {
	Sequence  uint64      `json:"sequence"`
	ProductID string      `json:"productId"`
	Field     ChangeField `json:"field"`
	Previous  int64       `json:"previous"`
	Current   int64       `json:"current"`
}

type PurchaseLine struct {
	LineNumber int    `json:"lineNumber"`
	ProductID  string `json:"productId"`
	Quantity   int    `json:"quantity"`
}

type PurchaseOrder struct {
	ID    string         `json:"id"`
	Lines []PurchaseLine `json:"lines"`
}

type PurchaseLineInput struct {
	ProductID string
	Quantity  int
}

type PurchaseReference struct {
	OrderID    string `json:"orderId"`
	LineNumber int    `json:"lineNumber"`
}

type DeletionCheck struct {
	Sequence       uint64 `json:"sequence"`
	ProductID      string `json:"productId"`
	ReferenceCount int    `json:"referenceCount"`
	Referenced     bool   `json:"referenced"`
}

type Catalog struct {
	Categories []Category `json:"categories"`
	Products   []Product  `json:"products"`
}

var (
	ErrInvalidCategory      = errors.New("invalid category")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrInvalidProduct       = errors.New("invalid product")
	ErrProductNotFound      = errors.New("product not found")
	ErrInvalidPrice         = errors.New("invalid price")
	ErrInvalidStock         = errors.New("invalid stock")
	ErrInvalidPurchaseOrder = errors.New("invalid purchase order")
	ErrProductReferenced    = errors.New("product is referenced by a purchase order")
)
