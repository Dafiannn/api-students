package service

import (
	"strings"
	"api-students-v7/app/model"
	"api-students-v7/helper"
)

// ApplyPatch menerapkan perubahan parsial yang dikirim klien
func ApplyPatch(current model.Student, req model.PatchStudentRequest) model.Student {
	if req.NIM != nil {
		current.NIM = strings.TrimSpace(*req.NIM)
	}
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}
	if req.Grade != nil {
		current.Grade = *req.Grade
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

// IsEmptyPatch memeriksa apakah request PATCH tidak mengirim satu pun field
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// CanAccessStudent memeriksa hak akses ownership vs permission
func CanAccessStudent(
	current model.AuthUser,
	ownerID *int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if ownerID != nil && current.UserID == *ownerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}
