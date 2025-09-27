package application

import (
	"bytes"
	"context"
	"ecs/internal/domain/product"
	csvs "ecs/internal/infrastructure/csv"
	"ecs/internal/infrastructure/database"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	es8 "github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

type ProductService struct {
	db       *gorm.DB
	repo     product.Repository
	esClient *es8.Client
}

func NewProductService(db *gorm.DB) *ProductService {
	repo := database.NewGormProductRepository(db)

	var es *es8.Client
	var err error
	for i := 0; i < 10; i++ {
		es, err = es8.NewClient(es8.Config{
			Addresses: []string{"http://es:9200"},
		})
		if err == nil {
			break
		}
		fmt.Println("Elasticsearch not ready, retrying in 3s...")
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		panic(fmt.Sprintf("failed to init elasticsearch client: %v", err))
	}

	return &ProductService{
		db:       db,
		repo:     repo,
		esClient: es,
	}
}

func (s *ProductService) GenerateCSV(total int, filePath string) error {
	return csvs.GenerateCSV(filePath, total)
}

func (s *ProductService) SeedFromCSV(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open csv: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("cannot read csv: %w", err)
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}
		price, _ := strconv.ParseFloat(row[1], 64)
		p := product.Product{Name: row[0], Price: price, Category: row[2], Tags: row[3]}
		if err := s.repo.Save(&p); err != nil {
			return fmt.Errorf("failed to save row: %w", err)
		}
	}
	return nil
}

func (s *ProductService) InitProductIndexEcs() error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"Name": map[string]interface{}{
					"type": "text",
				},
				"Name_suggest": map[string]interface{}{
					"type": "completion",
				},
				"Category": map[string]interface{}{
					"type": "keyword",
				},
				"Price": map[string]interface{}{
					"type": "float",
				},
				"Tags": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	body, _ := json.Marshal(mapping)

	exists, err := s.esClient.Indices.Exists([]string{"products"})
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer exists.Body.Close()

	if exists.StatusCode == 200 {
		fmt.Println("Index 'products' sudah ada, skip pembuatan index.")
		return nil
	}

	res, err := s.esClient.Indices.Create(
		"products",
		s.esClient.Indices.Create.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to create index, status: %s", res.String())
	}

	fmt.Println("Index 'products' berhasil dibuat")
	return nil
}

func (s *ProductService) SyncToElasticSearch() error {
	products, err := s.repo.FindAll()
	if err != nil {
		return err
	}

	for _, p := range products {
		doc := map[string]interface{}{
			"ID":           p.ID,
			"Name":         p.Name,
			"Price":        p.Price,
			"Category":     p.Category,
			"Tags":         p.Tags,
			"Name_suggest": map[string]interface{}{"input": []string{p.Name}},
		}
		body, _ := json.Marshal(doc)
		res, err := s.esClient.Index(
			"products",
			bytes.NewReader(body),
			s.esClient.Index.WithDocumentID(fmt.Sprint(p.ID)),
			s.esClient.Index.WithContext(context.Background()),
		)
		if err != nil {
			return fmt.Errorf("failed to index product %d: %w", p.ID, err)
		}
		res.Body.Close()
	}

	fmt.Printf("%d products synces to Elasticsearch\n", len(products))
	return nil
}

func (s *ProductService) CreateProduct(name string, price float64, category, tags string) error {
	p := product.Product{Name: name, Price: price, Category: category, Tags: tags}
	if err := s.repo.Save(&p); err != nil {
		return err
	}

	body, _ := json.Marshal(p)
	_, err := s.esClient.Index(
		"products",
		bytes.NewReader(body),
		s.esClient.Index.WithDocumentID(fmt.Sprint(p.ID)),
	)
	return err
}

func (s *ProductService) Autocomplete(query, category string, tags []string, minPrice, maxPrice float64) ([]product.Product, error) {
	var results []product.Product

	// Build ES query
	esQuery := map[string]interface{}{
		"suggest": map[string]interface{}{
			"product-suggest": map[string]interface{}{
				"prefix": query,
				"completion": map[string]interface{}{
					"field": "Name_suggest", // field khusus untuk suggest
					"fuzzy": map[string]interface{}{
						"fuzziness": "AUTO",
					},
					"size": 10,
				},
			},
		},
		"query": map[string]interface{}{ // filter tambahan
			"bool": map[string]interface{}{
				"filter": []interface{}{},
			},
		},
	}

	filters := []interface{}{}
	if category != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"Category": category,
			},
		})
	}

	if len(tags) > 0 {
		filters = append(filters, map[string]interface{}{
			"terms": map[string]interface{}{
				"Tags": tags,
			},
		})
	}

	if minPrice > 0 || maxPrice > 0 {
		priceRange := map[string]interface{}{}
		if minPrice > 0 {
			priceRange["gte"] = minPrice
		}
		if maxPrice > 0 {
			priceRange["lte"] = maxPrice
		}
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"Price": priceRange,
			},
		})
	}

	esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = filters

	queryBody, _ := json.Marshal(esQuery)
	fmt.Printf("DEBUG: ES Query = %s\n", string(queryBody))

	res, err := s.esClient.Search(
		s.esClient.Search.WithContext(context.Background()),
		s.esClient.Search.WithIndex("products"),
		s.esClient.Search.WithBody(bytes.NewReader(queryBody)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var r struct {
		Suggest map[string][]struct {
			Options []struct {
				Source product.Product `json:"_source"`
			} `json:"options"`
		} `json:"suggest"`
	}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}

	for _, suggestion := range r.Suggest["product-suggest"] {
		for _, opt := range suggestion.Options {
			results = append(results, opt.Source)
		}
	}

	return results, nil
}
