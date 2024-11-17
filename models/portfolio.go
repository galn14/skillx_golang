package models

type Portfolio struct {
	ID          string `json:"id" gorm:"primaryKey"`
	UserID      string `json:"user_id" gorm:"index"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Photo       string `json:"photo"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	DateCreated string `json:"date_created"`
	DateEnd     string `json:"date_end"`
	IsPresent   bool   `json:"is_present"`
}
