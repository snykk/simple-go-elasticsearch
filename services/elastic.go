package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/snykk/simple-go-elasticsearch/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/joho/godotenv"
)

var es *elasticsearch.Client

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	caCertPath := os.Getenv("ELASTICSEARCH_CA_PATH")
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatalf("Error reading CA certificate: %v", err)
	}

	es, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{os.Getenv("ELASTICSEARCH_URL")},
		Username:  os.Getenv("ELASTICSEARCH_USERNAME"),
		Password:  os.Getenv("ELASTICSEARCH_PASSWORD"),
		CACert:    caCert,
	})
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error connecting to Elasticsearch: %v", err)
	}
	defer res.Body.Close()
	log.Println("Connected to Elasticsearch:", res)
}

func IndexProducts(products []models.Product) error {
	for _, product := range products {
		productData, err := json.Marshal(product)
		if err != nil {
			return fmt.Errorf("Error marshaling product: %v", err)
		}

		req := bytes.NewReader(productData)
		_, err = es.Index(
			"products", // Nama indeks
			req,
			es.Index.WithDocumentID(product.ID),
			es.Index.WithRefresh("true"),
		)
		if err != nil {
			return fmt.Errorf("Error indexing product: %v", err)
		}
	}
	return nil
}

func LoadProductsFromFile(filePath string) ([]models.Product, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error reading products file: %v", err)
	}

	var products []models.Product
	err = json.Unmarshal(data, &products)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling products: %v", err)
	}

	return products, nil
}

func SearchProducts(query string) ([]models.Product, error) {
	// Query Elasticsearch untuk pencarian produk
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"name": query,
			},
		},
	}

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling search query: %v", err)
	}

	req := bytes.NewReader(body)
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("products"),
		es.Search.WithBody(req),
	)
	if err != nil {
		return nil, fmt.Errorf("Error searching products: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Error searching products: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("Error decoding search response: %v", err)
	}

	var products []models.Product
	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) {
		product := hit.(map[string]interface{})["_source"].(map[string]interface{})
		products = append(products, models.Product{
			ID:          product["id"].(string),
			Name:        product["name"].(string),
			Description: product["description"].(string),
			Price:       product["price"].(float64),
			Stock:       int(product["stock"].(float64)),
			CreatedAt:   product["created_at"].(string),
		})
	}

	return products, nil
}
