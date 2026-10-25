package service

import (
	"strings"
	"api-students-v5/app/model"
)

func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}
	trimmedNIM := strings.TrimSpace(req.NIM)
	trimmedName := strings.TrimSpace(req.Name)

	if trimmedNIM == "" {
		errs["nim"] = "wajib diisi"
	} else if len(trimmedNIM) < 5 {
		errs["nim"] = "minimal 5 karakter"
	}
	if trimmedName == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "nilai harus berada di antara 0 dan 100"
	}
	return errs
}

func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}
	trimmedNIM := strings.TrimSpace(req.NIM)
	trimmedName := strings.TrimSpace(req.Name)

	if trimmedNIM == "" {
		errs["nim"] = "wajib diisi pada PUT"
	} else if len(trimmedNIM) < 5 {
		errs["nim"] = "minimal 5 karakter"
	}
	if trimmedName == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "nilai harus berada di antara 0 dan 100"
	}
	return errs
}

func ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.NIM != nil {
		trimmed := strings.TrimSpace(*req.NIM)
		if trimmed == "" {
			errs["nim"] = "tidak boleh kosong"
		} else if len(trimmed) < 5 {
			errs["nim"] = "minimal 5 karakter"
		} else {
			current.NIM = trimmed
		}
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = trimmed
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "nilai harus berada di antara 0 dan 100"
		} else {
			current.Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
