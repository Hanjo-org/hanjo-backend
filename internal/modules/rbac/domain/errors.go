package domain

import "errors"

var (
	ErrRoleNotFound           = errors.New("peran tidak ditemukan")
	ErrRoleCodeExists         = errors.New("kode peran sudah digunakan dalam organisasi ini")
	ErrSystemRoleProtected    = errors.New("peran bawaan sistem tidak boleh diubah atau dihapus")
	ErrDeveloperRoleProtected = errors.New("peran developer dilindungi dan tidak dapat dimodifikasi")
	ErrPermissionNotFound     = errors.New("izin tidak ditemukan")
	ErrInvalidRoleTier        = errors.New("tier peran tidak valid, harus ORGANIZATION, BRAND, atau OUTLET")
	ErrInvalidRoleCode        = errors.New("format kode peran tidak valid, harus huruf kapital, angka, dan underscore")
	ErrRoleInUse              = errors.New("peran tidak dapat dihapus karena masih digunakan oleh karyawan")
)
