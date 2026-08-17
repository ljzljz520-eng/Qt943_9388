package stationery

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type MemoryStore struct {
	mu                   sync.RWMutex
	categories           map[string]Category
	products             map[string]Product
	history              map[string][]ChangeRecord
	orders               map[string]PurchaseOrder
	deletionChecks       []DeletionCheck
	nextProductID        uint64
	nextImageID          uint64
	nextChangeSequence   uint64
	nextOrderID          uint64
	nextDeletionSequence uint64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		categories: make(map[string]Category),
		products:   make(map[string]Product),
		history:    make(map[string][]ChangeRecord),
		orders:     make(map[string]PurchaseOrder),
	}
}

func (s *MemoryStore) AddCategory(ctx context.Context, category Category) (Category, error) {
	if err := contextError(ctx); err != nil {
		return Category{}, err
	}
	category.ID = strings.TrimSpace(category.ID)
	category.Name = strings.TrimSpace(category.Name)
	category.ParentID = strings.TrimSpace(category.ParentID)
	if category.ID == "" || category.Name == "" || !validCategoryAxis(category.Axis) {
		return Category{}, ErrInvalidCategory
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.categories[category.ID]; exists {
		return Category{}, fmt.Errorf("%w: category %s already exists", ErrInvalidCategory, category.ID)
	}
	if category.ParentID != "" {
		parent, exists := s.categories[category.ParentID]
		if !exists {
			return Category{}, fmt.Errorf("%w: %s", ErrCategoryNotFound, category.ParentID)
		}
		if parent.Axis != category.Axis {
			return Category{}, fmt.Errorf("%w: parent axis differs", ErrInvalidCategory)
		}
	}
	s.categories[category.ID] = category
	return category, nil
}

func (s *MemoryStore) Category(ctx context.Context, id string) (Category, error) {
	if err := contextError(ctx); err != nil {
		return Category{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	category, ok := s.categories[id]
	if !ok {
		return Category{}, fmt.Errorf("%w: %s", ErrCategoryNotFound, id)
	}
	return category, nil
}

func (s *MemoryStore) ListCategories(ctx context.Context) ([]Category, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	categories := make([]Category, 0, len(s.categories))
	for _, category := range s.categories {
		categories = append(categories, category)
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].ID < categories[j].ID
	})
	return categories, nil
}

func (s *MemoryStore) AddProduct(ctx context.Context, input AddProductInput) (Product, error) {
	if err := contextError(ctx); err != nil {
		return Product{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	grade, ok := s.categories[input.GradeCategoryID]
	if !ok {
		return Product{}, fmt.Errorf("%w: %s", ErrCategoryNotFound, input.GradeCategoryID)
	}
	purpose, ok := s.categories[input.PurposeCategoryID]
	if !ok {
		return Product{}, fmt.Errorf("%w: %s", ErrCategoryNotFound, input.PurposeCategoryID)
	}
	if grade.Axis != CategoryAxisGrade || purpose.Axis != CategoryAxisPurpose {
		return Product{}, fmt.Errorf("%w: category axes", ErrInvalidProduct)
	}

	s.nextProductID++
	productID := fmt.Sprintf("P%04d", s.nextProductID)
	images := make([]ProductImage, len(input.Images))
	for i, image := range input.Images {
		s.nextImageID++
		images[i] = ProductImage{
			ID:        fmt.Sprintf("IMG%04d", s.nextImageID),
			ProductID: productID,
			URI:       strings.TrimSpace(image.URI),
			Alt:       strings.TrimSpace(image.Alt),
			Position:  i + 1,
		}
	}
	product := Product{
		ID:                productID,
		Name:              strings.TrimSpace(input.Name),
		Kind:              input.Kind,
		GradeCategoryID:   input.GradeCategoryID,
		PurposeCategoryID: input.PurposeCategoryID,
		PriceCents:        input.PriceCents,
		Stock:             input.Stock,
		Images:            images,
	}
	s.products[product.ID] = product
	return cloneProduct(product), nil
}

func (s *MemoryStore) Product(ctx context.Context, id string) (Product, error) {
	if err := contextError(ctx); err != nil {
		return Product{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	product, ok := s.products[id]
	if !ok {
		return Product{}, fmt.Errorf("%w: %s", ErrProductNotFound, id)
	}
	return cloneProduct(product), nil
}

func (s *MemoryStore) ListProducts(ctx context.Context) ([]Product, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	products := make([]Product, 0, len(s.products))
	for _, product := range s.products {
		products = append(products, cloneProduct(product))
	}
	sort.Slice(products, func(i, j int) bool {
		return products[i].ID < products[j].ID
	})
	return products, nil
}

func (s *MemoryStore) UpdatePrice(ctx context.Context, id string, priceCents int64) (Product, error) {
	if err := contextError(ctx); err != nil {
		return Product{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	product, ok := s.products[id]
	if !ok {
		return Product{}, fmt.Errorf("%w: %s", ErrProductNotFound, id)
	}
	previous := product.PriceCents
	product.PriceCents = priceCents
	s.products[id] = product
	s.appendChangeLocked(id, ChangeFieldPrice, previous, priceCents)
	return cloneProduct(product), nil
}

func (s *MemoryStore) UpdateStock(ctx context.Context, id string, stock int) (Product, error) {
	if err := contextError(ctx); err != nil {
		return Product{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	product, ok := s.products[id]
	if !ok {
		return Product{}, fmt.Errorf("%w: %s", ErrProductNotFound, id)
	}
	previous := product.Stock
	product.Stock = stock
	s.products[id] = product
	s.appendChangeLocked(id, ChangeFieldStock, int64(previous), int64(stock))
	return cloneProduct(product), nil
}

func (s *MemoryStore) History(ctx context.Context, productID string) ([]ChangeRecord, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.products[productID]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrProductNotFound, productID)
	}
	records := append([]ChangeRecord(nil), s.history[productID]...)
	return records, nil
}

func (s *MemoryStore) CreatePurchaseOrder(ctx context.Context, lines []PurchaseLineInput) (PurchaseOrder, error) {
	if err := contextError(ctx); err != nil {
		return PurchaseOrder{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(lines) == 0 {
		return PurchaseOrder{}, ErrInvalidPurchaseOrder
	}
	orderLines := make([]PurchaseLine, len(lines))
	for i, line := range lines {
		if line.Quantity <= 0 {
			return PurchaseOrder{}, ErrInvalidPurchaseOrder
		}
		if _, ok := s.products[line.ProductID]; !ok {
			return PurchaseOrder{}, fmt.Errorf("%w: %s", ErrProductNotFound, line.ProductID)
		}
		orderLines[i] = PurchaseLine{LineNumber: i + 1, ProductID: line.ProductID, Quantity: line.Quantity}
	}
	s.nextOrderID++
	order := PurchaseOrder{ID: fmt.Sprintf("PO%04d", s.nextOrderID), Lines: orderLines}
	s.orders[order.ID] = order
	return cloneOrder(order), nil
}

func (s *MemoryStore) PurchaseOrder(ctx context.Context, id string) (PurchaseOrder, error) {
	if err := contextError(ctx); err != nil {
		return PurchaseOrder{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[id]
	if !ok {
		return PurchaseOrder{}, fmt.Errorf("%w: order %s", ErrInvalidPurchaseOrder, id)
	}
	for _, line := range order.Lines {
		if _, ok := s.products[line.ProductID]; !ok {
			return PurchaseOrder{}, fmt.Errorf("%w: missing product %s", ErrInvalidPurchaseOrder, line.ProductID)
		}
	}
	return cloneOrder(order), nil
}

func (s *MemoryStore) ListPurchaseLinesByProduct(ctx context.Context, productID string) ([]PurchaseReference, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	references := make([]PurchaseReference, 0)
	for orderID, order := range s.orders {
		for _, line := range order.Lines {
			if line.ProductID == productID {
				references = append(references, PurchaseReference{OrderID: orderID, LineNumber: line.LineNumber})
			}
		}
	}
	sort.Slice(references, func(i, j int) bool {
		if references[i].OrderID == references[j].OrderID {
			return references[i].LineNumber < references[j].LineNumber
		}
		return references[i].OrderID < references[j].OrderID
	})
	return references, nil
}

func (s *MemoryStore) RecordDeletionCheck(ctx context.Context, check DeletionCheck) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextDeletionSequence++
	check.Sequence = s.nextDeletionSequence
	s.deletionChecks = append(s.deletionChecks, check)
	return nil
}

func (s *MemoryStore) DeleteProduct(ctx context.Context, id string) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.products[id]; !ok {
		return fmt.Errorf("%w: %s", ErrProductNotFound, id)
	}
	delete(s.products, id)
	return nil
}

func (s *MemoryStore) appendChangeLocked(productID string, field ChangeField, previous, current int64) {
	s.nextChangeSequence++
	s.history[productID] = append(s.history[productID], ChangeRecord{
		Sequence:  s.nextChangeSequence,
		ProductID: productID,
		Field:     field,
		Previous:  previous,
		Current:   current,
	})
}

func validCategoryAxis(axis CategoryAxis) bool {
	return axis == CategoryAxisGrade || axis == CategoryAxisPurpose
}

func contextError(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func cloneProduct(product Product) Product {
	product.Images = append([]ProductImage(nil), product.Images...)
	return product
}

func cloneOrder(order PurchaseOrder) PurchaseOrder {
	order.Lines = append([]PurchaseLine(nil), order.Lines...)
	return order
}
