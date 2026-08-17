package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"campus-stationery/internal/stationery"
)

func main() {
	ctx := context.Background()
	service := stationery.NewService(stationery.NewMemoryStore())
	if _, err := stationery.LoadCampusFixture(ctx, service); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	catalog, err := service.Catalog(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(catalog); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
