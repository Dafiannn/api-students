package main

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var students []Student
var nextID = 1

func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

func isNIMExists(nim string, excludeID int) bool {
	for _, s := range students {
		if s.ID != excludeID && strings.EqualFold(s.NIM, nim) {
			return true
		}
	}
	return false
}

// GET /api/v1/students
func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)

	// 1. Penyaringan (Filtering)
	hasil := []Student{}
	for _, s := range students {
		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue
		}
		if q.Search != "" {
			keyword := strings.ToLower(q.Search)
			nameMatch := strings.Contains(strings.ToLower(s.Name), keyword)
			nimMatch := strings.Contains(strings.ToLower(s.NIM), keyword)
			if !nameMatch && !nimMatch {
				continue
			}
		}
		hasil = append(hasil, s)
	}

	// 2. Pengurutan (Sorting)
	sort.SliceStable(hasil, func(i, j int) bool {
		var lebihKecil bool
		switch q.Sort {
		case "nim":
			lebihKecil = hasil[i].NIM < hasil[j].NIM
		case "name":
			lebihKecil = strings.ToLower(hasil[i].Name) < strings.ToLower(hasil[j].Name)
		case "grade":
			lebihKecil = hasil[i].Grade < hasil[j].Grade
		default:
			lebihKecil = hasil[i].ID < hasil[j].ID
		}
		if q.Order == "desc" {
			return !lebihKecil
		}
		return lebihKecil
	})

	// 3. Paginasi (Pagination)
	total := len(hasil)
	totalPages := 0
	if q.Limit > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	mulai := (q.Page - 1) * q.Limit
	if mulai > total {
		mulai = total
	}
	akhir := mulai + q.Limit
	if akhir > total {
		akhir = total
	}

	return okList(c, "daftar mahasiswa berhasil diambil", hasil[mulai:akhir], &Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GET /api/v1/students/:id
func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	idx := findStudentIndex(id)
	if idx == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	return ok(c, "data mahasiswa ditemukan", students[idx])
}

// POST /api/v1/students
func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := map[string]string{}
	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}
	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "nilai harus berada di antara 0 dan 100"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Cek keunikan NIM (409 Conflict)
	if isNIMExists(req.NIM, 0) {
		return fail(c, fiber.StatusConflict, "NIM sudah terdaftar dalam sistem")
	}

	baru := Student{
		ID:        nextID,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	students = append(students, baru)
	nextID++

	return created(c, "mahasiswa berhasil ditambahkan", baru, "/api/v1/students/"+strconv.Itoa(baru.ID))
}

// PUT /api/v1/students/:id
func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	idx := findStudentIndex(id)
	if idx == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	errs := map[string]string{}
	if req.NIM == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if req.Name == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "nilai harus berada di antara 0 dan 100"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	if isNIMExists(req.NIM, id) {
		return fail(c, fiber.StatusConflict, "NIM sudah terdaftar untuk mahasiswa lain")
	}

	students[idx].NIM = req.NIM
	students[idx].Name = req.Name
	students[idx].Grade = req.Grade
	students[idx].IsActive = req.IsActive

	return ok(c, "data mahasiswa berhasil diganti seluruhnya", students[idx])
}

// PATCH /api/v1/students/:id
func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	idx := findStudentIndex(id)
	if idx == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	errs := map[string]string{}
	if req.NIM != nil {
		trimmed := strings.TrimSpace(*req.NIM)
		if trimmed == "" {
			errs["nim"] = "tidak boleh kosong"
		} else if isNIMExists(trimmed, id) {
			return fail(c, fiber.StatusConflict, "NIM sudah terdaftar untuk mahasiswa lain")
		} else {
			students[idx].NIM = trimmed
		}
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			students[idx].Name = trimmed
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "nilai harus berada di antara 0 dan 100"
		} else {
			students[idx].Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		students[idx].IsActive = *req.IsActive
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	return ok(c, "data mahasiswa berhasil diperbarui sebagian", students[idx])
}

// DELETE /api/v1/students/:id
func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	idx := findStudentIndex(id)
	if idx == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	students = append(students[:idx], students[idx+1:]...)
	return noContent(c)
}
