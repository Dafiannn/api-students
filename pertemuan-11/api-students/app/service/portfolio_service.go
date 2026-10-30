package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"api-students-v11/app/model"
	"api-students-v11/app/repository"
	"api-students-v11/helper"
)

type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
	perms      *helper.PermissionSet
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
	perms *helper.PermissionSet,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students, perms: perms}
}

// GET /students/:id/portfolio
func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// Pemeriksaan lintas database: Mahasiswa wajib ada di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
		}
		return helper.Internal(err)
	}

	portfolio, err := s.portfolios.GetByStudentID(ctx, studentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("portofolio mahasiswa belum dibuat")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diambil", portfolio)
}

// PUT /students/:id/portfolio
func (s *PortfolioService) Upsert(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// Pemeriksaan lintas database: Tolak bila mahasiswa tidak ada di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
		}
		return helper.Internal(err)
	}

	var req model.UpsertPortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	portfolio, err := s.portfolios.Upsert(ctx, studentID, req)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil disimpan", portfolio)
}

// POST /students/:id/portfolio/projects
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan di PostgreSQL")
		}
		return helper.Internal(err)
	}

	var req model.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	portfolio, err := s.portfolios.AddProject(ctx, studentID, req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return helper.NotFound("portofolio mahasiswa belum dibuat, buat portofolio terlebih dahulu")
		case errors.Is(err, repository.ErrProjectLimit):
			return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT", "jumlah proyek sudah mencapai batas maksimal 20 proyek")
		default:
			return helper.Internal(err)
		}
	}

	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan ke portofolio", portfolio)
}

// DELETE /students/:id/portfolio/projects/:projectId
func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	projectID := c.Params("projectId")
	if projectID == "" {
		return helper.BadRequest("id proyek tidak boleh kosong")
	}

	portfolio, err := s.portfolios.DeleteProject(ctx, studentID, projectID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidID):
			return helper.BadRequest("id proyek tidak valid")
		case errors.Is(err, repository.ErrNotFound):
			return helper.NotFound("portofolio tidak ditemukan")
		case errors.Is(err, repository.ErrProjectNotFound):
			return helper.NotFound("proyek tidak ditemukan dalam portofolio")
		default:
			return helper.Internal(err)
		}
	}

	return helper.Success(c, fiber.StatusOK, "proyek berhasil dihapus dari portofolio", portfolio)
}

// GET /portfolios/stats/technologies
func (s *PortfolioService) StatsTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	stats, err := s.portfolios.TopTechnologies(ctx)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "statistik 10 teknologi paling banyak digunakan", stats)
}
