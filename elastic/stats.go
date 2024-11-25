package elastic

import (
	"context"
	"encoding/json"
	"fmt"
)

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
