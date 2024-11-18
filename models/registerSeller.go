package models

import "time"

type registerSeller struct {
	UID             string    `json:"uid"`
	Name            string    `json:"name"`         // nama mahasiswa
	Status          string    `json:"status"`       // "pending", "denied", "accepted"
	Email           string    `json:"email"`        // email mahasiswa
	Organization    string    `json:"organization"` // asal kampus
	Major           string    `json:"major"`        //jurusannya
	PhotoURL        string    `json:"photo_url"`    //foto id card mahasiswa
	Verified        bool      `json:"verified"`
	GraduationMonth string    `json:"graduation_month,omitempty"`
	GraduationYear  int       `json:"graduation_year,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
