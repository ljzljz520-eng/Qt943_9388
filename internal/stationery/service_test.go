package stationery_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"campus-stationery/internal/stationery"
)

func TestCatalogContainsCategoryHierarchyImagesAndEveryProductKind(t *testing.T) {
	service, fixture := newFixture(t)

	catalog, err := service.Catalog(context.Background())
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	if len(catalog.Categories) != 7 {
		t.Errorf("category count = %d, want 7", len(catalog.Categories))
	}
	if len(catalog.Products) != 4 {
		t.Errorf("product count = %d, want 4", len(catalog.Products))
	}
	marker := fixture.Products[stationery.ProductKindMarker]
	if marker.GradeCategoryID != "grade-primary" || marker.PurposeCategoryID != "purpose-art" {
		t.Errorf("marker categories = %q/%q", marker.GradeCategoryID, marker.PurposeCategoryID)
	}
	if len(marker.Images) != 2 || marker.Images[0].ProductID != marker.ID {
		t.Errorf("marker images = %+v", marker.Images)
	}
}

func TestPriceAndStockChangesRemainInSequence(t *testing.T) {
	service, fixture := newFixture(t)
	product := fixture.Products[stationery.ProductKindNotebook]

	if _, err := service.ChangePrice(context.Background(), product.ID, 725); err != nil {
		t.Fatalf("change price: %v", err)
	}
	if _, err := service.ChangeStock(context.Background(), product.ID, 64); err != nil {
		t.Fatalf("change stock: %v", err)
	}
	history, err := service.ChangeHistory(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	want := []stationery.ChangeRecord{
		{Sequence: 1, ProductID: product.ID, Field: stationery.ChangeFieldPrice, Previous: 650, Current: 725},
		{Sequence: 2, ProductID: product.ID, Field: stationery.ChangeFieldStock, Previous: 80, Current: 64},
	}
	if !reflect.DeepEqual(history, want) {
		t.Errorf("history = %+v, want %+v", history, want)
	}
}

func TestUnreferencedFolderCanBeDeleted(t *testing.T) {
	service, fixture := newFixture(t)
	product := fixture.Products[stationery.ProductKindFolder]

	if err := service.DeleteProduct(context.Background(), product.ID); err != nil {
		t.Fatalf("delete folder: %v", err)
	}
	if _, err := service.Product(context.Background(), product.ID); !errors.Is(err, stationery.ErrProductNotFound) {
		t.Errorf("product lookup error = %v, want product not found", err)
	}
}

func TestReferencedMarkerDeletionIsRejectedAndPurchaseOrderRemainsValid(t *testing.T) {
	service, fixture := newFixture(t)
	marker := fixture.Products[stationery.ProductKindMarker]
	order, err := service.CreatePurchaseOrder(context.Background(), []stationery.PurchaseLineInput{{ProductID: marker.ID, Quantity: 12}})
	if err != nil {
		t.Fatalf("create purchase order: %v", err)
	}

	err = service.DeleteProduct(context.Background(), marker.ID)
	if !errors.Is(err, stationery.ErrProductReferenced) {
		t.Errorf("delete error = %v, want product referenced", err)
	}
	if _, err := service.Product(context.Background(), marker.ID); err != nil {
		t.Errorf("marker lookup: %v", err)
	}
	if _, err := service.PurchaseOrder(context.Background(), order.ID); err != nil {
		t.Errorf("purchase order lookup: %v", err)
	}
}

func newFixture(t *testing.T) (*stationery.Service, stationery.Fixture) {
	t.Helper()
	service := stationery.NewService(stationery.NewMemoryStore())
	fixture, err := stationery.LoadCampusFixture(context.Background(), service)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	return service, fixture
}
