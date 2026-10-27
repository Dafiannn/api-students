package model

import "time"

// Portfolio merepresentasikan portofolio mahasiswa di MongoDB (collection student_portfolios).
type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type Project struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Tech  []string `json:"tech"`
	URL   string   `json:"url"`
	Year  int      `json:"year"`
}

type Certificate struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Issuer string    `json:"issuer"`
	Date   time.Time `json:"date"`
}

type UpsertPortfolioRequest struct {
	Headline     string                  `json:"headline" validate:"required,min=3,max=150"`
	Bio          string                  `json:"bio" validate:"required,min=10,max=2000"`
	Projects     []CreateProjectRequest  `json:"projects,omitempty" validate:"omitempty,max=20,dive"`
	Certificates []CreateCertRequest     `json:"certificates,omitempty" validate:"omitempty,max=20,dive"`
}

type CreateProjectRequest struct {
	Title string   `json:"title" validate:"required,min=3,max=100"`
	Tech  []string `json:"tech" validate:"required,min=1,max=10,dive,min=1,max=30"`
	URL   string   `json:"url" validate:"omitempty,url"`
	Year  int      `json:"year" validate:"required,gte=2000,lte=2100"`
}

type CreateCertRequest struct {
	Name   string    `json:"name" validate:"required,min=3,max=100"`
	Issuer string    `json:"issuer" validate:"required,min=2,max=100"`
	Date   time.Time `json:"date" validate:"required"`
}

type TechStat struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}
