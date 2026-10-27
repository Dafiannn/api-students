package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"api-students-v10/app/model"
)

var (
	ErrProjectLimit = errors.New("jumlah proyek sudah mencapai batas maksimal (20 proyek)")
	ErrProjectNotFound = errors.New("proyek tidak ditemukan dalam portofolio")
)

const MaxProjectsPerPortfolio = 20

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	GetByStudentID(ctx context.Context, studentID int) (model.Portfolio, error)
	Upsert(ctx context.Context, studentID int, req model.UpsertPortfolioRequest) (model.Portfolio, error)
	AddProject(ctx context.Context, studentID int, p model.CreateProjectRequest) (model.Portfolio, error)
	DeleteProject(ctx context.Context, studentID int, projectID string) (model.Portfolio, error)
	TopTechnologies(ctx context.Context) ([]model.TechStat, error)
}

type portfolioDocument struct {
	ID           bson.ObjectID         `bson:"_id,omitempty"`
	StudentID    int                   `bson:"student_id"`
	Headline     string                `bson:"headline"`
	Bio          string                `bson:"bio"`
	Projects     []projectDocument     `bson:"projects"`
	Certificates []certificateDocument `bson:"certificates"`
	CreatedAt    time.Time             `bson:"created_at"`
	UpdatedAt    time.Time             `bson:"updated_at"`
}

type projectDocument struct {
	ID    bson.ObjectID `bson:"_id"`
	Title string        `bson:"title"`
	Tech  []string      `bson:"tech"`
	URL   string        `bson:"url"`
	Year  int           `bson:"year"`
}

type certificateDocument struct {
	ID     bson.ObjectID `bson:"_id"`
	Name   string        `bson:"name"`
	Issuer string        `bson:"issuer"`
	Date   time.Time     `bson:"date"`
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		tech := p.Tech
		if tech == nil {
			tech = []string{}
		}
		projects = append(projects, model.Project{
			ID:    p.ID.Hex(),
			Title: p.Title,
			Tech:  tech,
			URL:   p.URL,
			Year:  p.Year,
		})
	}

	certs := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certs = append(certs, model.Certificate{
			ID:     c.ID.Hex(),
			Name:   c.Name,
			Issuer: c.Issuer,
			Date:   c.Date,
		})
	}

	return model.Portfolio{
		ID:           d.ID.Hex(),
		StudentID:    d.StudentID,
		Headline:     d.Headline,
		Bio:          d.Bio,
		Projects:     projects,
		Certificates: certs,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

type portfolioMongoRepository struct {
	coll *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("student_portfolio_unique"),
		},
		{
			Keys:    bson.D{{Key: "projects.tech", Value: 1}},
			Options: options.Index().SetName("portfolio_project_tech"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) GetByStudentID(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

func (r *portfolioMongoRepository) Upsert(ctx context.Context, studentID int, req model.UpsertPortfolioRequest) (model.Portfolio, error) {
	t := now()
	projects := make([]projectDocument, 0, len(req.Projects))
	for _, p := range req.Projects {
		projects = append(projects, projectDocument{
			ID:    bson.NewObjectID(),
			Title: p.Title,
			Tech:  p.Tech,
			URL:   p.URL,
			Year:  p.Year,
		})
	}

	certs := make([]certificateDocument, 0, len(req.Certificates))
	for _, c := range req.Certificates {
		certs = append(certs, certificateDocument{
			ID:     bson.NewObjectID(),
			Name:   c.Name,
			Issuer: c.Issuer,
			Date:   c.Date,
		})
	}

	update := bson.M{
		"$set": bson.M{
			"headline":     req.Headline,
			"bio":          req.Bio,
			"projects":     projects,
			"certificates": certs,
			"updated_at":   t,
		},
		"$setOnInsert": bson.M{
			"student_id": studentID,
			"created_at": t,
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var doc portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"student_id": studentID}, update, opts).Decode(&doc)
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("menyimpan portofolio: %w", err)
	}
	return doc.toModel(), nil
}

func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, p model.CreateProjectRequest) (model.Portfolio, error) {
	projDoc := projectDocument{
		ID:    bson.NewObjectID(),
		Title: p.Title,
		Tech:  p.Tech,
		URL:   p.URL,
		Year:  p.Year,
	}

	limitField := fmt.Sprintf("projects.%d", MaxProjectsPerPortfolio-1)
	filter := bson.M{
		"student_id": studentID,
		limitField:   bson.M{"$exists": false},
	}

	update := bson.M{
		"$push": bson.M{"projects": projDoc},
		"$set":  bson.M{"updated_at": now()},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var doc portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err == nil {
		return doc.toModel(), nil
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	// Cari tahu alasan gagal: apakah portofolio tidak ada atau kuota proyek penuh
	count, countErr := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if countErr != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", countErr)
	}
	if count == 0 {
		return model.Portfolio{}, ErrNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

func (r *portfolioMongoRepository) DeleteProject(ctx context.Context, studentID int, projectID string) (model.Portfolio, error) {
	oid, err := parseID(projectID)
	if err != nil {
		return model.Portfolio{}, err
	}

	filter := bson.M{
		"student_id":   studentID,
		"projects._id": oid,
	}

	update := bson.M{
		"$pull": bson.M{"projects": bson.M{"_id": oid}},
		"$set":  bson.M{"updated_at": now()},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var doc portfolioDocument
	err = r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err == nil {
		return doc.toModel(), nil
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		// Cari tahu apakah portofolio-nya yang tidak ada atau proyeknya
		count, _ := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
		if count == 0 {
			return model.Portfolio{}, ErrNotFound
		}
		return model.Portfolio{}, ErrProjectNotFound
	}
	return model.Portfolio{}, fmt.Errorf("menghapus proyek: %w", err)
}

func (r *portfolioMongoRepository) TopTechnologies(ctx context.Context) ([]model.TechStat, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.tech"}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.tech",
			"total": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "total", Value: -1},
			{Key: "_id", Value: 1},
		}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("agregasi teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var raw []struct {
		Tech  string `bson:"_id"`
		Total int    `bson:"total"`
	}
	if err := cursor.All(ctx, &raw); err != nil {
		return nil, fmt.Errorf("membaca hasil agregasi teknologi: %w", err)
	}

	results := make([]model.TechStat, 0, len(raw))
	for _, item := range raw {
		results = append(results, model.TechStat{
			Technology: item.Tech,
			Total:      item.Total,
		})
	}
	return results, nil
}
