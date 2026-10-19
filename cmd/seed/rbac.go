package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PermissionSeed defines an atomic permission to be seeded.
type PermissionSeed struct {
	Code        string
	Group       string
	Name        string
	Description string
}

// RoleSeed defines a system default role to be seeded.
type RoleSeed struct {
	Code        string
	Name        string
	Description string
	Tier        string // 'ORGANIZATION', 'BRAND', 'OUTLET'
	Permissions []string
}

// MasterPermissionsCatalog lists all official FnB ERP & POS system permissions.
var MasterPermissionsCatalog = []PermissionSeed{
	// 1. User & Staff Management
	{Code: "user:staff:read", Group: "user", Name: "Lihat Daftar Karyawan", Description: "Melihat data profil karyawan dan penugasan cabang"},
	{Code: "user:staff:create", Group: "user", Name: "Tambah Karyawan Baru", Description: "Mendaftarkan staf baru dan menentukan role/cabang"},
	{Code: "user:staff:update", Group: "user", Name: "Ubah Data Karyawan", Description: "Memperbarui profil staf, role, atau mutasi cabang"},
	{Code: "user:staff:delete", Group: "user", Name: "Nonaktifkan Karyawan", Description: "Menandai status staf nonaktif atau menghapus akun"},

	// 2. RBAC & Custom Roles
	{Code: "rbac:role:read", Group: "rbac", Name: "Lihat Daftar Peran & Izin", Description: "Melihat role bawaan sistem dan role kustom organisasi"},
	{Code: "rbac:role:manage", Group: "rbac", Name: "Kelola Custom Role", Description: "Membuat atau mengubah role kustom dan permission matrix"},

	// 3. Organization & Outlets
	{Code: "org:brand:read", Group: "org", Name: "Lihat Daftar Brand", Description: "Melihat daftar brand di bawah payung holding"},
	{Code: "org:brand:manage", Group: "org", Name: "Kelola Brand", Description: "Menambah atau mengubah informasi brand konsep FnB"},
	{Code: "org:outlet:read", Group: "org", Name: "Lihat Daftar Outlet", Description: "Melihat daftar outlet cabang fisik dan kontaknya"},
	{Code: "org:outlet:manage", Group: "org", Name: "Kelola Outlet", Description: "Menambah atau mengonfigurasi outlet dan jam operasional"},

	// 4. Menu & Catalog
	{Code: "catalog:item:read", Group: "catalog", Name: "Lihat Menu & Kategori", Description: "Melihat katalog produk menu, varian, dan harga"},
	{Code: "catalog:item:manage", Group: "catalog", Name: "Kelola Menu & Produk", Description: "Membuat, mengubah, atau menghapus item menu"},
	{Code: "catalog:price:override", Group: "catalog", Name: "Penyesuaian Harga Outlet", Description: "Menyesuaikan harga khusus outlet tertentu"},
	{Code: "catalog:recipe:view_cost", Group: "catalog", Name: "Lihat HPP & Biaya Resep", Description: "Melihat kalkulasi HPP/COGS dan formula bahan rahasia"},

	// 5. Point of Sale (POS)
	{Code: "pos:order:read", Group: "pos", Name: "Lihat Transaksi Penjualan", Description: "Melihat daftar pesanan dan riwayat tiket kasir"},
	{Code: "pos:order:create", Group: "pos", Name: "Buat Pesanan Baru", Description: "Mencatat pesanan baru, tiket dapur, dan cetak bill"},
	{Code: "pos:order:payment", Group: "pos", Name: "Terima Pembayaran", Description: "Memproses pembayaran tunai, QRIS, EDC, atau kartu"},
	{Code: "pos:order:void", Group: "pos", Name: "Void Pesanan / Item", Description: "Membatalkan pesanan terkirim (Memerlukan PIN Supervisor)"},
	{Code: "pos:order:discount_apply", Group: "pos", Name: "Berikan Diskon Manual", Description: "Memberikan potongan manual (Memerlukan PIN Supervisor)"},
	{Code: "pos:shift:open_close", Group: "pos", Name: "Buka / Tutup Shift Kas", Description: "Input modal awal laci dan rekonsiliasi kas kasir"},
	{Code: "pos:drawer:open", Group: "pos", Name: "Buka Laci Kasir Manual", Description: "Membuka laci uang tanpa transaksi (Memerlukan PIN Supervisor)"},

	// 6. Inventory & Purchasing
	{Code: "inventory:stock:read", Group: "inventory", Name: "Lihat Saldo Stok Bahan", Description: "Melihat ketersediaan stok bahan di cabang/gudang"},
	{Code: "inventory:opname:manage", Group: "inventory", Name: "Kelola Stok Opname", Description: "Input penyesuaian fisik opname stok berkala"},
	{Code: "inventory:transfer:request", Group: "inventory", Name: "Minta Transfer Stok", Description: "Mengajukan mutasi bahan antar-cabang atau ke gudang pusat"},
	{Code: "inventory:transfer:approve", Group: "inventory", Name: "Setujui Mutasi Stok", Description: "Otorisasi pengiriman mutasi stok bahan baku"},
	{Code: "inventory:po:create", Group: "inventory", Name: "Buat Purchase Order", Description: "Mengajukan pesanan pembelian ke supplier"},
	{Code: "inventory:po:approve", Group: "inventory", Name: "Persetujuan PO Pembelian", Description: "Otorisasi PO pembelian skala brand/holding"},

	// 7. Finance & Reports
	{Code: "finance:report:sales", Group: "finance", Name: "Laporan Penjualan", Description: "Melihat ringkasan omzet harian dan bulanan"},
	{Code: "finance:report:pnl", Group: "finance", Name: "Laporan Laba Rugi", Description: "Melihat laporan laba rugi konsolidasi & HPP operasional"},
	{Code: "finance:payment_method:manage", Group: "finance", Name: "Kelola Metode Pembayaran", Description: "Mengatur channel pembayaran, bank, dan EDC"},
}

// allPermissionCodes returns a slice of all permission codes in the catalog.
func allPermissionCodes() []string {
	codes := make([]string, len(MasterPermissionsCatalog))
	for i, p := range MasterPermissionsCatalog {
		codes[i] = p.Code
	}
	return codes
}

// SystemRolesCatalog defines the 9 system default roles and their default permission mappings.
var SystemRolesCatalog = []RoleSeed{
	{
		Code:        "DEVELOPER",
		Name:        "Platform Developer",
		Description: "Platform Super Admin - Hidden from lookups with full authorization bypass",
		Tier:        "ORGANIZATION",
		Permissions: allPermissionCodes(), // All permissions granted
	},
	{
		Code:        "OWNER",
		Name:        "Pemilik Bisnis (Owner)",
		Description: "Akses penuh manajemen holding, konfigurasi organisasi, dan keuangan konsolidasi",
		Tier:        "ORGANIZATION",
		Permissions: allPermissionCodes(), // All business permissions granted
	},
	{
		Code:        "FINANCE_HQ",
		Name:        "Finance & Akuntansi HQ",
		Description: "Konsolidasi keuangan, rekonsiliasi pembayaran, laporan laba rugi, dan approval PO",
		Tier:        "ORGANIZATION",
		Permissions: []string{
			"user:staff:read",
			"org:brand:read",
			"org:outlet:read",
			"catalog:item:read",
			"catalog:recipe:view_cost",
			"pos:order:read",
			"inventory:stock:read",
			"inventory:po:approve",
			"finance:report:sales",
			"finance:report:pnl",
			"finance:payment_method:manage",
		},
	},
	{
		Code:        "CENTRAL_PURCHASING",
		Name:        "Purchasing & Logistik Pusat",
		Description: "Pengadaan bahan baku holding, manajemen gudang pusat, dan transfer stok",
		Tier:        "ORGANIZATION",
		Permissions: []string{
			"org:brand:read",
			"org:outlet:read",
			"catalog:item:read",
			"catalog:recipe:view_cost",
			"inventory:stock:read",
			"inventory:transfer:request",
			"inventory:transfer:approve",
			"inventory:po:create",
			"inventory:po:approve",
		},
	},
	{
		Code:        "BRAND_MANAGER",
		Name:        "Manajer Brand",
		Description: "Bertanggung jawab atas 1 brand: menu, formula resep (BOM), dan promo pemasaran",
		Tier:        "BRAND",
		Permissions: []string{
			"org:brand:read",
			"org:outlet:read",
			"catalog:item:read",
			"catalog:item:manage",
			"catalog:price:override",
			"catalog:recipe:view_cost",
			"pos:order:read",
			"inventory:stock:read",
			"finance:report:sales",
		},
	},
	{
		Code:        "BRANCH_MANAGER",
		Name:        "Manajer Cabang",
		Description: "Bertanggung jawab atas 1 cabang: operasional, shift, stok opname, dan PIN supervisor",
		Tier:        "OUTLET",
		Permissions: []string{
			"user:staff:read",
			"user:staff:create",
			"user:staff:update",
			"org:brand:read",
			"org:outlet:read",
			"catalog:item:read",
			"catalog:price:override",
			"pos:order:read",
			"pos:order:create",
			"pos:order:payment",
			"pos:order:void",
			"pos:order:discount_apply",
			"pos:shift:open_close",
			"pos:drawer:open",
			"inventory:stock:read",
			"inventory:opname:manage",
			"inventory:transfer:request",
			"finance:report:sales",
		},
	},
	{
		Code:        "CASHIER",
		Name:        "Kasir POS",
		Description: "Operasional kasir garis depan: transaksi kasir, pembayaran, dan buka/tutup shift",
		Tier:        "OUTLET",
		Permissions: []string{
			"catalog:item:read",
			"pos:order:read",
			"pos:order:create",
			"pos:order:payment",
			"pos:shift:open_close",
		},
	},
	{
		Code:        "BARISTA",
		Name:        "Barista / Tim Dapur",
		Description: "Kitchen Display System (KDS): melihat dan memproses tiket pesanan",
		Tier:        "OUTLET",
		Permissions: []string{
			"catalog:item:read",
			"pos:order:read",
		},
	},
	{
		Code:        "WAITER",
		Name:        "Pelayan Meja (Waiter)",
		Description: "Pemesanan meja tamu (Table Order) dan pengecekan status ketersediaan meja",
		Tier:        "OUTLET",
		Permissions: []string{
			"catalog:item:read",
			"pos:order:read",
			"pos:order:create",
		},
	},
}

// Internal seeding database models
type permDB struct {
	ID          string `gorm:"column:id;primaryKey"`
	Code        string `gorm:"column:code"`
	Group       string `gorm:"column:group"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	CreatedAt   time.Time
}

func (permDB) TableName() string { return "permissions" }

type roleDB struct {
	ID             string  `gorm:"column:id;primaryKey"`
	OrganizationID *string `gorm:"column:organization_id"`
	Code           string  `gorm:"column:code"`
	Name           string  `gorm:"column:name"`
	Description    string  `gorm:"column:description"`
	Tier           string  `gorm:"column:tier"`
	IsSystem       bool    `gorm:"column:is_system"`
	Status         string  `gorm:"column:status"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (roleDB) TableName() string { return "roles" }

type rolePermDB struct {
	RoleID       string `gorm:"column:role_id;primaryKey"`
	PermissionID string `gorm:"column:permission_id;primaryKey"`
	CreatedAt    time.Time
}

func (rolePermDB) TableName() string { return "role_permissions" }

// seedRBAC performs idempotent seeding of master permissions, 9 system roles, and role-permission mappings.
// Returns a map of role code -> role ID for user seeding.
func seedRBAC(ctx context.Context, db *gorm.DB, log *zap.Logger) (map[string]string, error) {
	fmt.Println("\n🛡️  Seeding RBAC Master Permissions & System Roles...")

	// 1. Seed Permissions
	permMap := make(map[string]string, len(MasterPermissionsCatalog))
	newPermsCount := 0

	for _, p := range MasterPermissionsCatalog {
		var existing permDB
		err := db.WithContext(ctx).First(&existing, "code = ?", p.Code).Error
		if err != nil {
			newID := uuid.NewString()
			created := permDB{
				ID:          newID,
				Code:        p.Code,
				Group:       p.Group,
				Name:        p.Name,
				Description: p.Description,
				CreatedAt:   time.Now().UTC(),
			}
			if err := db.WithContext(ctx).Create(&created).Error; err != nil {
				return nil, fmt.Errorf("failed to insert permission %s: %w", p.Code, err)
			}
			permMap[p.Code] = newID
			newPermsCount++
		} else {
			permMap[p.Code] = existing.ID
		}
	}
	fmt.Printf("✅ Master Permissions synced (%d total, %d newly created)\n", len(permMap), newPermsCount)

	// 2. Seed 9 System Roles
	rolesMap := make(map[string]string, len(SystemRolesCatalog))
	newRolesCount := 0

	for _, r := range SystemRolesCatalog {
		var existing roleDB
		err := db.WithContext(ctx).
			First(&existing, "code = ? AND organization_id IS NULL AND deleted_at IS NULL", r.Code).
			Error
		if err != nil {
			newID := uuid.NewString()
			now := time.Now().UTC()
			created := roleDB{
				ID:             newID,
				OrganizationID: nil, // System default roles are global (organization_id = NULL)
				Code:           r.Code,
				Name:           r.Name,
				Description:    r.Description,
				Tier:           r.Tier,
				IsSystem:       true,
				Status:         "ACTIVE",
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := db.WithContext(ctx).Create(&created).Error; err != nil {
				return nil, fmt.Errorf("failed to insert role %s: %w", r.Code, err)
			}
			rolesMap[r.Code] = newID
			newRolesCount++
			fmt.Printf("   ➕ Role created: %-18s [%s]\n", r.Code, r.Tier)
		} else {
			rolesMap[r.Code] = existing.ID
		}
	}
	fmt.Printf("✅ System Roles synced (%d total, %d newly created)\n", len(rolesMap), newRolesCount)

	// 3. Seed Role-Permission Mappings
	totalMappings := 0
	for _, r := range SystemRolesCatalog {
		roleID, ok := rolesMap[r.Code]
		if !ok {
			continue
		}

		for _, permCode := range r.Permissions {
			permID, ok := permMap[permCode]
			if !ok {
				continue
			}

			var existing rolePermDB
			err := db.WithContext(ctx).
				First(&existing, "role_id = ? AND permission_id = ?", roleID, permID).
				Error
			if err != nil {
				mapping := rolePermDB{
					RoleID:       roleID,
					PermissionID: permID,
					CreatedAt:    time.Now().UTC(),
				}
				if err := db.WithContext(ctx).Create(&mapping).Error; err != nil {
					return nil, fmt.Errorf("failed to map role %s to perm %s: %w", r.Code, permCode, err)
				}
				totalMappings++
			}
		}
	}
	fmt.Printf("✅ Role-Permission mappings synced successfully\n")

	return rolesMap, nil
}
