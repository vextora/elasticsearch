package main

import (
	"ecs/internal/application"
	"ecs/internal/domain/product"
	"ecs/internal/infrastructure/database"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/app/main.go [run|seed|sync]")
		os.Exit(1)
	}

	action := os.Args[1]
	csvFile := "/app/db/products.csv"

	switch action {
	case "all":
		service := initService()

		fmt.Println("===> Running full flow: generate, seed, sync")

		if err := service.GenerateCSV(200, csvFile); err != nil {
			fmt.Println("Generate CSV error:", err)
		}
		if err := service.SeedFromCSV(csvFile); err != nil {
			fmt.Println("Seed error:", err)
		}
		if err := service.CreateProduct("Sample Product", 99.99, "electronics", "trending,new,handmade"); err != nil {
			fmt.Println("CreateProduct error:", err)
		}
		if err := service.SyncToElasticSearch(); err != nil {
			fmt.Println("Sync error:", err)
		}
	case "seed":
		service := initService()

		fmt.Println("===> Seeding database from CSV...")

		if err := service.GenerateCSV(200, csvFile); err != nil {
			fmt.Println("Generate CSV error:", err)
		}

		fmt.Println("===> Seed to database")
		if err := service.SeedFromCSV(csvFile); err != nil {
			fmt.Println("Seed error:", err)
		}
	case "sync":
		service := initService()

		fmt.Println("===> Syncing data to Elasticsearch...")

		if err := service.SyncToElasticSearch(); err != nil {
			fmt.Println("Sync error:", err)
		}
	default:
		fmt.Println("Unknown command: ", action)
		os.Exit(1)
	}
}

func initService() *application.ProductService {
	fmt.Println("===> Connection to database...")
	db := database.NewPostgresDB()

	fmt.Println("===> Migrating database schema...")
	if err := db.AutoMigrate(&product.Product{}); err != nil {
		panic(fmt.Sprintf("gagal auto-migrate: %v", err))
	}
	fmt.Println("Migration selesai.")

	service := application.NewProductService(db)
	return service
}
