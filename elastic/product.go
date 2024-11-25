package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/snykk/simple-go-elasticsearch/models"
)

func IndexProducts(products []models.Product) error {
	for _, product := range products {
		productData, err := json.Marshal(map[string]interface{}{
			"id":          product.ID,
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"stock":       product.Stock,
			"category":    product.Category,
			"created_at":  product.CreatedAt,
			"suggest": map[string]interface{}{
				"input":  []string{product.Name, product.Category}, // Kata kunci untuk fitur suggest
				"weight": 10,                                       // Bobot relevansi
			},
		})
		if err != nil {
			return fmt.Errorf("error marshaling product: %v", err)
		}

		req := bytes.NewReader(productData)
		_, err = es.Index(
			"products", // Nama index
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

func SearchProductsWithPaginationSortingAndFilter(query, page, size, sort, priceMin, priceMax, stockMin, stockMax, category, createdAtMin, createdAtMax string) ([]models.Product, error) {
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
							"name": map[string]interface{}{
								"query":     query,
								"fuzziness": "AUTO", // Mendukung pencarian fuzzy
							},
						},
					},
				},
				"filter": []map[string]interface{}{},
			},
		},
		"sort": []map[string]interface{}{
			{
				sort: map[string]interface{}{
					"order": "asc",
				},
			},
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"name":        map[string]interface{}{},
				"description": map[string]interface{}{},
			},
		},
	}

	if category != "" {
		searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"term": map[string]interface{}{
					"category": category,
				},
			},
		)
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

	if createdAtMin != "" && createdAtMax != "" {
		searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"range": map[string]interface{}{
					"created_at": map[string]interface{}{
						"gte": createdAtMin,
						"lte": createdAtMax,
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
			Category:    product["category"].(string),
			CreatedAt:   product["created_at"].(string),
		})
	}

	return products, nil
}

func AggregateProductData() (map[string]interface{}, error) {
	aggregationQuery := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"average_price": map[string]interface{}{
				"avg": map[string]interface{}{
					"field": "price",
				},
			},
			"min_price": map[string]interface{}{
				"min": map[string]interface{}{
					"field": "price",
				},
			},
			"max_price": map[string]interface{}{
				"max": map[string]interface{}{
					"field": "price",
				},
			},
		},
	}

	body, err := json.Marshal(aggregationQuery)
	if err != nil {
		return nil, fmt.Errorf("error marshaling aggregation query: %v", err)
	}

	req := bytes.NewReader(body)
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("products"),
		es.Search.WithBody(req),
	)
	if err != nil {
		return nil, fmt.Errorf("error performing aggregation: %v", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding aggregation response: %v", err)
	}

	aggregations := result["aggregations"].(map[string]interface{})
	averagePrice := aggregations["average_price"].(map[string]interface{})["value"]
	minPrice := aggregations["min_price"].(map[string]interface{})["value"]
	maxPrice := aggregations["max_price"].(map[string]interface{})["value"]

	response := map[string]interface{}{
		"average_price": averagePrice,
		"min_price":     minPrice,
		"max_price":     maxPrice,
	}

	return response, nil
}

func SuggestProductData(query string) ([]string, error) {
	suggestQuery := map[string]interface{}{
		"suggest": map[string]interface{}{
			"product-suggest": map[string]interface{}{
				"prefix": query,
				"completion": map[string]interface{}{
					"field": "suggest",
					"size":  5, // Max 5 suggestions
				},
			},
		},
	}

	body, err := json.Marshal(suggestQuery)
	if err != nil {
		return nil, fmt.Errorf("error marshaling suggest query: %v", err)
	}

	req := bytes.NewReader(body)
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("products"),
		es.Search.WithBody(req),
	)
	if err != nil {
		return nil, fmt.Errorf("error performing suggest query: %v", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding suggest response: %v", err)
	}

	suggestions := []string{}
	for _, suggestion := range result["suggest"].(map[string]interface{})["product-suggest"].([]interface{}) {
		options := suggestion.(map[string]interface{})["options"].([]interface{})
		for _, option := range options {
			suggestionText := option.(map[string]interface{})["text"].(string)
			suggestions = append(suggestions, suggestionText)
		}
	}

	return suggestions, nil
}

func UpdateProducts(updates []models.Product) error {
	for _, update := range updates {
		updateData, err := json.Marshal(update)
		if err != nil {
			return fmt.Errorf("error marshaling update data: %v", err)
		}

		req := bytes.NewReader(updateData)
		_, err = es.Update(
			"products",
			update.ID,
			req,
			es.Update.WithRefresh("true"),
		)
		if err != nil {
			return fmt.Errorf("error updating product with ID %s: %v", update.ID, err)
		}
	}
	return nil
}
