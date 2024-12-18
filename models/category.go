package models

type Category struct {
	IdCategory string `json:"id_category,omitempty"`
	Title      string `json:"title"`
	PhotoUrl   string `json:"photo_url"`
	IdMajor    string `json:"id_major,omitempty"`
}
