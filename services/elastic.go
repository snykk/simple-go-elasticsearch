package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/snykk/simple-go-elasticsearch/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/joho/godotenv"
)

var es *elasticsearch.Client

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	caCertPath := os.Getenv("ELASTICSEARCH_CA_PATH")
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatalf("error reading CA certificate: %v", err)
	}

	es, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{os.Getenv("ELASTICSEARCH_URL")},
		Username:  os.Getenv("ELASTICSEARCH_USERNAME"),
		Password:  os.Getenv("ELASTICSEARCH_PASSWORD"),
		CACert:    caCert,
	})
	if err != nil {
		log.Fatalf("error creating Elasticsearch client: %v", err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("error connecting to Elasticsearch: %v", err)
	}
	defer res.Body.Close()
	log.Println("Connected to Elasticsearch:", res)
}

func IndexProducts(products []models.Product) error {
	for _, product := range products {
		productData, err := json.Marshal(product)
		if err != nil {
			return fmt.Errorf("error marshaling product: %v", err)
		}

		req := bytes.NewReader(productData)
		_, err = es.Index(
			"products", // Index name
			req,
			es.Index.WithDocumentID(product.ID),
			es.Index.WithRefresh("true"),
		)
		if err != nil {
			return fmt.Errorf("error indexing product: %v", err)
		}
	}
	return nil
}

func LoadProductsFromFile(filePath string) ([]models.Product, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading products file: %v", err)
	}

	var products []models.Product
	err = json.Unmarshal(data, &products)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling products: %v", err)
	}

	return products, nil
}

func SearchProductsWithPaginationSortingAndFilter(query, page, size, sort, priceMin, priceMax, stockMin, stockMax string) ([]models.Product, error) {
	pageInt, _ := strconv.Atoi(page)
	sizeInt, _ := strconv.Atoi(size)
	offset := (pageInt - 1) * sizeInt

	searchQuery := map[string]interface{}{
		"from": offset,
		"size": sizeInt,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							"name": query,
						},
					},
				},
				"filter": []map[string]interface{}{},
			},
		},
		"sort": []map[string]interface{}{
			{
				sort: map[string]interface{}{
					"order": "asc", // Sorting ascending, bisa diubah jadi "desc" untuk descending
				},
			},
		},
	}

	if priceMin != "" && priceMax != "" {
		searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"range": map[string]interface{}{
					"price": map[string]interface{}{
						"gte": priceMin,
						"lte": priceMax,
					},
				},
			},
		)
	}

	if stockMin != "" && stockMax != "" {
		searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"range": map[string]interface{}{
					"stock": map[string]interface{}{
						"gte": stockMin,
						"lte": stockMax,
					},
				},
			},
		)
	}

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("error marshaling search query: %v", err)
	}

	req := bytes.NewReader(body)
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("products"),
		es.Search.WithBody(req),
	)
	if err != nil {
		return nil, fmt.Errorf("error searching products: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching products: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding search response: %v", err)
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

func GetIndexStats() (map[string]interface{}, error) {
	res, err := es.Indices.Stats(
		es.Indices.Stats.WithContext(context.Background()),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting index stats: %v", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding stats response: %v", err)
	}

	return result, nil
}
