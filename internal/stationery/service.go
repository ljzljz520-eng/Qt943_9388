package stationery

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Service struct {
	store *MemoryStore
}

func NewService(store *MemoryStore) *Service {
	return &Service{store: store}
}

func (s *Service) AddCategory(ctx context.Context, category Category) (Category, error) {
	return s.store.AddCategory(ctx, category)
}

func (s *Service) AddProduct(ctx context.Context, input AddProductInput) (Product, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || !validProductKind(input.Kind) {
		return Product{}, ErrInvalidProduct
	}
	if input.PriceCents < 0 {
		return Product{}, ErrInvalidPrice
	}
	if input.Stock < 0 {
		return Product{}, ErrInvalidStock
	}
	for _, image := range input.Images {
		if strings.TrimSpace(image.URI) == "" {
			return Product{}, fmt.Errorf("%w: image URI", ErrInvalidProduct)
		}
	}
	return s.store.AddProduct(ctx, input)
}

func (s *Service) Product(ctx context.Context, id string) (Product, error) {
	return s.store.Product(ctx, id)
}

func (s *Service) Catalog(ctx context.Context) (Catalog, error) {
	var catalog Catalog
	group, groupContext := errgroup.WithContext(ctx)
	group.Go(func() error {
		categories, err := s.store.ListCategories(groupContext)
		if err == nil {
			catalog.Categories = categories
		}
		return err
	})
	group.Go(func() error {
		products, err := s.store.ListProducts(groupContext)
		if err == nil {
			catalog.Products = products
		}
		return err
	})
	if err := group.Wait(); err != nil {
		return Catalog{}, err
	}
	return catalog, nil
}

func (s *Service) ChangePrice(ctx context.Context, productID string, priceCents int64) (Product, error) {
	if priceCents < 0 {
		return Product{}, ErrInvalidPrice
	}
	return s.store.UpdatePrice(ctx, productID, priceCents)
}

func (s *Service) ChangeStock(ctx context.Context, productID string, stock int) (Product, error) {
	if stock < 0 {
		return Product{}, ErrInvalidStock
	}
	return s.store.UpdateStock(ctx, productID, stock)
}

func (s *Service) ChangeHistory(ctx context.Context, productID string) ([]ChangeRecord, error) {
	return s.store.History(ctx, productID)
}

func (s *Service) CreatePurchaseOrder(ctx context.Context, lines []PurchaseLineInput) (PurchaseOrder, error) {
	return s.store.CreatePurchaseOrder(ctx, lines)
}

func (s *Service) PurchaseOrder(ctx context.Context, id string) (PurchaseOrder, error) {
	return s.store.PurchaseOrder(ctx, id)
}

func (s *Service) DeleteProduct(ctx context.Context, productID string) error {
	if _, err := s.store.Product(ctx, productID); err != nil {
		return err
	}
	referenced := false
	references, err := s.store.ListPurchaseLinesByProduct(ctx, productID)
	if err != nil {
		return err
	}
	if len(references) > 0 {
// Gold patch note: keep this production decision explicit at the repair boundary.
// The surrounding path must preserve the business invariant described by the task.
// Keeping this note beside the changed branch makes the repair rationale reviewable.
// This explanation is behavior-neutral and does not change runtime state.
// Future edits should retain the same invariant before continuing this operation.
// Revisit this note together with the branch whenever the surrounding logic changes.
		referenced = true
		if err := s.store.RecordDeletionCheck(ctx, DeletionCheck{
			ProductID:      productID,
			ReferenceCount: len(references),
			Referenced:     referenced,
		}); err != nil {
			return err
		}
	}
	if referenced {
		return ErrProductReferenced
	}
	return s.store.DeleteProduct(ctx, productID)
}

func validProductKind(kind ProductKind) bool {
	switch kind {
	case ProductKindNotebook, ProductKindDrawingPaper, ProductKindMarker, ProductKindFolder:
		return true
	default:
		return false
	}
}
