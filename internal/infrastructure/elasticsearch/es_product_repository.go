package elasticsearch

import (
	"bytes"
	"context"
	"ecs/internal/domain/product"
	"encoding/json"
	"fmt"
	"log"

	es8 "github.com/elastic/go-elasticsearch/v8"
)

type ESProductRepository struct {
	client *es8.Client
}

func NewESProductRepository() *ESProductRepository {
	cfg := es8.Config{
		Addresses: []string{"http://elasticsearch:9200"},
	}

	client, err := es8.NewClient(cfg)
	if err != nil {
		log.Fatalf("failed to create elastic client: %v", err)
	}
	return &ESProductRepository{client: client}
}

func (e *ESProductRepository) Index(p *product.Product) error {
	body, _ := json.Marshal(p)
	res, err := e.client.Index(
		"products",
		bytes.NewReader(body),
		e.client.Index.WithDocumentID(fmt.Sprint(p.ID)),
		e.client.Index.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Elasticsearch error: %s", res.String())
	}
	return nil
}
