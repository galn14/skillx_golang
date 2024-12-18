package models

type Service struct {
	IdService    string `json:"id_service,omitempty"`
	TitleService string `json:"title_service"`
	IconUrl      string `json:"icon_url"`
	IdCategory   string `json:"id_category,omitempty"`
}
