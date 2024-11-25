package models

type SuggestField struct {
	Input  []string `json:"input"`  // Kata kunci suggest
	Weight int      `json:"weight"` // Bobot relevansi
}

type Product struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       float64      `json:"price"`
	Stock       int          `json:"stock"`
	Category    string       `json:"category"`
	CreatedAt   string       `json:"created_at"`
	Suggest     SuggestField `json:"suggest"`
}
