package service

import (
	"strings"
	"api-students-v6/app/model"
	"api-students-v6/helper"
)

// CanAccessStudent memutuskan apakah seseorang berhak mengakses/mengubah data mahasiswa.
// Jalur 1: Kepemilikan (ownership) — user tersebut adalah pemilik data (owner_id).
// Jalur 2: Permission — role-nya memiliki permission :any (misal admin/staff).
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

// ValidateAssignRole memvalidasi perubahan role user lain
func ValidateAssignRole(
	current model.AuthUser,
	targetID int,
	req model.AssignRoleRequest,
	perms *helper.PermissionSet,
) map[string]string {
	errs := map[string]string{}
	role := strings.TrimSpace(req.Role)
	if role == "" {
		errs["role"] = "wajib diisi"
		return errs
	}
	if !perms.IsKnownRole(role) {
		errs["role"] = "role tidak dikenal, pilih salah satu dari: " +
			strings.Join(perms.KnownRoles(), ", ")
	}
	// Aturan proteksi: tidak boleh mengubah role diri sendiri
	if current.UserID == targetID {
		errs["role"] = "tidak boleh mengubah role diri sendiri"
	}
	return errs
}
