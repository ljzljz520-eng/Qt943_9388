package stationery

import "context"

type Fixture struct {
	Categories map[string]Category
	Products   map[ProductKind]Product
}

func LoadCampusFixture(ctx context.Context, service *Service) (Fixture, error) {
	categories := []Category{
		{ID: "grade", Name: "All grades", Axis: CategoryAxisGrade},
		{ID: "grade-primary", Name: "Primary school", Axis: CategoryAxisGrade, ParentID: "grade"},
		{ID: "grade-middle", Name: "Middle school", Axis: CategoryAxisGrade, ParentID: "grade"},
		{ID: "purpose", Name: "All purposes", Axis: CategoryAxisPurpose},
		{ID: "purpose-writing", Name: "Writing", Axis: CategoryAxisPurpose, ParentID: "purpose"},
		{ID: "purpose-art", Name: "Art", Axis: CategoryAxisPurpose, ParentID: "purpose"},
		{ID: "purpose-filing", Name: "Filing", Axis: CategoryAxisPurpose, ParentID: "purpose"},
	}
	fixture := Fixture{Categories: make(map[string]Category), Products: make(map[ProductKind]Product)}
	for _, category := range categories {
		stored, err := service.AddCategory(ctx, category)
		if err != nil {
			return Fixture{}, err
		}
		fixture.Categories[stored.ID] = stored
	}
	products := []AddProductInput{
		{Name: "Grid Notebook", Kind: ProductKindNotebook, GradeCategoryID: "grade-middle", PurposeCategoryID: "purpose-writing", PriceCents: 650, Stock: 80, Images: []ImageInput{{URI: "images/grid-notebook.webp", Alt: "Grid notebook cover"}}},
		{Name: "A3 Drawing Paper", Kind: ProductKindDrawingPaper, GradeCategoryID: "grade-primary", PurposeCategoryID: "purpose-art", PriceCents: 1200, Stock: 45, Images: []ImageInput{{URI: "images/a3-paper.webp", Alt: "A3 drawing paper pack"}}},
		{Name: "Washable Markers", Kind: ProductKindMarker, GradeCategoryID: "grade-primary", PurposeCategoryID: "purpose-art", PriceCents: 1800, Stock: 30, Images: []ImageInput{{URI: "images/markers-front.webp", Alt: "Marker set"}, {URI: "images/markers-colors.webp", Alt: "Marker colors"}}},
		{Name: "Tabbed Folder", Kind: ProductKindFolder, GradeCategoryID: "grade-middle", PurposeCategoryID: "purpose-filing", PriceCents: 900, Stock: 55, Images: []ImageInput{{URI: "images/tabbed-folder.webp", Alt: "Tabbed folder"}}},
	}
	for _, input := range products {
		product, err := service.AddProduct(ctx, input)
		if err != nil {
			return Fixture{}, err
		}
		fixture.Products[product.Kind] = product
	}
	return fixture, nil
}
