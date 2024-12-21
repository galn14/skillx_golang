package models

import "time"

type Product struct {
	UID         string    `json:"uid"`
	NameProduct string    `json:"nameProduct"`
	Description string    `json:"description"`
	PhotoURL    []string  `json:"photo_url"` // Ganti string dengan []string
	Price       string    `json:"price"`
	Major       string    `json:"major"` //jurusannya
	IdCategory  string    `json:"idCategory"`
	IdService   string    `json:"idService"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
