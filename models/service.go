package models

type Service struct {
	IdService    string `json:"idService"`
	TitleService string `json:"titleService"`
	IconUrl      string `json:"iconUrl"`
	Link         string `json:"link"`
}
