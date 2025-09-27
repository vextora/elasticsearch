package main

import (
	"ecs/internal/application"
	"ecs/internal/domain/product"
	"ecs/internal/infrastructure/database"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	service := initService()
	startServer(service)

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

func startServer(service *application.ProductService) {
	csvFile := "/app/db/products.csv"

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/internal/seed", func(c *gin.Context) {
		if os.Getenv("ENABLE_SEED") != "true" {
			c.JSON(http.StatusForbidden, gin.H{"error": "seeding is disabled"})
			return
		}

		if err := service.GenerateCSV(200, csvFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Generate CSV error"})
			return
		}

		if err := service.SeedFromCSV(csvFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Seed error: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "seed completed"})
	})

	r.POST("/internal/ecs", func(c *gin.Context) {
		if err := service.InitProductIndexEcs(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Create index error: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "create index done"})
	})

	r.POST("/internal/sync", func(c *gin.Context) {
		if os.Getenv("ENABLE_SEED") != "true" {
			c.JSON(http.StatusForbidden, gin.H{"error": "seeding is disabled"})
			return
		}

		if err := service.SyncToElasticSearch(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Sync error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "sync done"})
	})

	r.GET("/autocomplete", func(c *gin.Context) {
		query := c.Query("q")
		category := c.Query("category")
		minPrice := 0.0
		maxPrice := 0.0

		results, err := service.Autocomplete(query, category, []string{}, minPrice, maxPrice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if results == nil {
			results = []product.Product{}
		}

		c.JSON(http.StatusOK, gin.H{
			"count": len(results),
			"data":  results,
		})
	})

	fmt.Println("===> Starting API server on :8080")
	if err := r.Run(":8080"); err != nil {
		panic(fmt.Sprintf("failed to start server: %v", err))
	}
}
