# Planning & Brainstorming: Sistem RBAC & Multi-Tenant Scoping

Dokumen ini merangkum arsitektur, desain skema database, taksonomi permission, strategi performa, dan rencana implementasi untuk sistem **Role-Based Access Control (RBAC)** serta **Multi-Tenant Hierarchy Scoping** pada Hanjo ERP (Enterprise Multi-Tenant FnB ERP & POS Backend).

---

## 1. Latar Belakang & Kebutuhan Spesifik FnB ERP

Pada aplikasi Food & Beverage (FnB) skala enterprise, sistem otorisasi tidak cukup hanya mengandalkan pengecekan role sederhana (*"Admin vs User"*). Terdapat 4 karakteristik fundamental industri FnB yang mendasari desain arsitektur ini:

1. **Pemisahan Dimensi Hak Akses (*Permission*) vs Wilayah Data (*Tenant Scope*)**:
   * **Permission ("What can you do?")**: Menentukan aksi teknis apa yang diizinkan (misal: buat pesanan, void item, lihat HPP/margin resep, cetak laporan laba rugi).
   * **Scope ("Where can you do it?")**: Menentukan di mana data tersebut boleh diakses berdasarkan hierarki 3-Tier (`Organization` -> `Brand` -> `Outlet`).
2. **Dual Client & Persona (Backoffice Web vs POS Frontline)**:
   * **Backoffice Web (HQ / Manajemen)**: Sesi login panjang (Email + Password), navigasi luas, membutuhkan hak akses granular per modul (Akuntansi, Procurement, HR/Staff, Katalog Menu).
   * **Frontline POS / KDS (Tablet / Layar Sentuh)**: Operasional serba cepat di outlet fisik, login cepat menggunakan 6-digit PIN, pergantian kasir/barista dalam hitungan detik.
3. **Supervisor / Manager PIN Override (Fitur Vital POS)**:
   * Kasir garis depan seringkali tidak memiliki izin untuk melakukan *Void pesanan* atau memberikan *Diskon Manual*.
   * Saat antrean padat, kasir tidak boleh logout. POS membutuhkan modal cepat: *"Masukkan PIN Supervisor"* untuk memvalidasi otorisasi di tempat, dan mencatat siapa manajer yang menyetujui aksi tersebut dalam audit log.
4. **Fleksibilitas Custom Role**:
   * Setiap bisnis FnB memiliki struktur organisasi berbeda. Ada restoran yang memiliki posisi *"Head Barista"*, *"Captain Waiter"*, *"Store Supervisor"*, atau *"Internal Auditor"*. Sistem harus mengizinkan Owner/Holding untuk membuat role baru dan mencentang izin spesifik dari Backoffice.

---

## 2. Arsitektur Model: Hybrid Access Control (Scoped RBAC + Contextual Policy Guards)

Sistem menggunakan model **Hybrid Access Control** (kombinasi Scoped Dynamic RBAC, Tenant Scoping, dan Contextual Policy Guards) sebagaimana didokumentasikan secara lengkap pada [access-control-hybrid-model.md](../knowledge/access-control-hybrid-model.md):

```text
┌─────────────────────────┐
│       Permissions       │ (System-wide Atomic Actions)
│  e.g. pos:order:void    │
└────────────┬────────────┘
             │ (Many-to-Many)
             ▼
┌─────────────────────────┐
│          Roles          │ (System Default & Custom Org Roles)
│   e.g. CASHIER, OWNER   │
└────────────┬────────────┘
             │ (1 Role per User)
             ▼
┌─────────────────────────┐
│          Users          │ (Identitas Karyawan)
│  role_id, org_id        │
│  brand_id, outlet_id    │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│     Tenant Boundary     │ (Filter Query Otomatis)
│   Tier 1: Organization  │
│   Tier 2: Brand         │
│   Tier 3: Outlet        │
└─────────────────────────┘
```

### Klasifikasi Tier pada Role

Setiap role memiliki atribut `tier` yang membatasi level penugasan user:

| Role Tier | Penugasan User | Contoh Role Bawaan | Lingkup Akses Data |
| :--- | :--- | :--- | :--- |
| **`ORGANIZATION`** | `organization_id` wajib<br>`brand_id = NULL`<br>`outlet_id = NULL` | `OWNER`<br>`FINANCE_HQ`<br>`CENTRAL_PURCHASING` | Akses penuh konsolidasi ke seluruh brand dan outlet dalam satu organisasi/holding. |
| **`BRAND`** | `organization_id` wajib<br>`brand_id` wajib<br>`outlet_id = NULL` | `BRAND_MANAGER`<br>`RND_CHEF` | Terbatas pada master katalog, resep, dan promo untuk brand yang ditugaskan di seluruh cabangnya. |
| **`OUTLET`** | `organization_id` wajib<br>`brand_id` wajib<br>`outlet_id` wajib | `BRANCH_MANAGER`<br>`CASHIER`<br>`BARISTA`<br>`KITCHEN`<br>`WAITER` | Terkunci pada 1 outlet fisik tempat user bertugas (transaksi kasir, shift kas, meja, stok harian cabang). |

---

## 3. Desain Skema Database (DDL PostgreSQL)

### 3.1. Tabel `permissions`
Menyimpan seluruh daftar izin atomik yang disediakan oleh sistem (di-seed oleh developer).

```sql
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) NOT NULL UNIQUE,       -- format: module:resource:action (e.g. 'pos:order:void')
    "group" VARCHAR(50) NOT NULL,            -- e.g. 'pos', 'inventory', 'menu', 'user', 'finance'
    name VARCHAR(150) NOT NULL,             -- Label UI (e.g. 'Void Pesanan POS')
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_permissions_group ON permissions("group");
```

### 3.2. Tabel `roles`
Menyimpan role sistem default serta custom role yang dibuat oleh Owner organisasi.

```sql
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NULL REFERENCES organizations(id) ON DELETE CASCADE, 
    -- NULL = System Default Role (bawaan sistem, berlaku global)
    -- NOT NULL = Custom Role milik organisasi tertentu
    code VARCHAR(50) NOT NULL,               -- e.g. 'OWNER', 'BRANCH_MANAGER', 'CASHIER'
    name VARCHAR(100) NOT NULL,              -- Display Name: 'Kasir POS'
    description TEXT,
    tier VARCHAR(20) NOT NULL,               -- 'ORGANIZATION', 'BRAND', 'OUTLET'
    is_system BOOLEAN NOT NULL DEFAULT FALSE,-- Role bawaan sistem tidak boleh dihapus
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

-- Partial unique index: code harus unik per organisasi (atau unik di system global)
CREATE UNIQUE INDEX uq_roles_org_code ON roles(organization_id, code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uq_roles_system_code ON roles(code) WHERE organization_id IS NULL AND deleted_at IS NULL;
CREATE INDEX idx_roles_org_id ON roles(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_roles_tier ON roles(tier) WHERE deleted_at IS NULL;
```

### 3.3. Tabel `role_permissions`
Menghubungkan role dengan permission yang dimilikinya (Junction Table).

```sql
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_perm_id ON role_permissions(permission_id);
```

### 3.4. Relasi pada Tabel `users`
Menambahkan kolom `role_id` pada tabel `users` yang sudah ada:

```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES roles(id) ON DELETE RESTRICT;
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id) WHERE deleted_at IS NULL;
```

### 3.5. Dukungan Multi-Outlet / Area Manager (Tabel `user_outlets`)
Untuk karyawan yang membawahi lebih dari 1 outlet (misal: Area Manager membawahi 3 cabang tertentu):

```sql
CREATE TABLE IF NOT EXISTS user_outlets (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    outlet_id UUID NOT NULL REFERENCES outlets(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, outlet_id)
);

CREATE INDEX idx_user_outlets_user_id ON user_outlets(user_id);
CREATE INDEX idx_user_outlets_outlet_id ON user_outlets(outlet_id);
```

---

## 4. Katalog Master Permissions (Slug Taxonomy)

Format standar penamaan: `<module>:<resource>:<action>`

### A. Modul User & Role Management (`user`, `rbac`)
* `user:staff:read` - Melihat daftar karyawan & profil.
* `user:staff:create` - Mendaftarkan staf baru.
* `user:staff:update` - Mengubah data staf, role, atau cabang.
* `user:staff:delete` - Menonaktifkan staf.
* `rbac:role:read` - Melihat daftar role & permissions.
* `rbac:role:manage` - Membuat atau mengubah custom role & permissions.

### B. Modul Organization & Outlets (`organization`)
* `org:brand:read` - Melihat daftar brand.
* `org:brand:manage` - Menambah/mengubah brand dan logo.
* `org:outlet:read` - Melihat daftar outlet.
* `org:outlet:manage` - Menambah/mengubah outlet, jam operasional, dan konfigurasi pajak.

### C. Modul Menu & Catalog (`catalog`)
* `catalog:item:read` - Melihat daftar kategori & produk menu.
* `catalog:item:manage` - Membuat, mengubah, dan menghapus menu.
* `catalog:price:override` - Menyesuaikan harga khusus outlet tertentu.
* `catalog:recipe:view_cost` - Melihat biaya HPP/COGS dan formula resep rahasia (Restricted).

### D. Modul Point of Sale (`pos`)
* `pos:order:read` - Melihat daftar transaksi penjualan outlet.
* `pos:order:create` - Mencatat pesanan baru dan mencetak bill/tiket dapur.
* `pos:order:payment` - Menyelesaikan pembayaran (Tunai, QRIS, Kartu, EDC).
* `pos:order:void` - Membatalkan pesanan/item yang sudah terkirim ke dapur (*Memerlukan Manager PIN*).
* `pos:order:discount_apply` - Memberikan diskon khusus/manual pada transaksi.
* `pos:shift:open_close` - Membuka modal awal laci kasir dan closing shift kasir.
* `pos:drawer:open` - Membuka laci uang tanpa transaksi kasir (*Memerlukan Manager PIN*).

### E. Modul Inventory & Purchasing (`inventory`)
* `inventory:stock:read` - Melihat saldo stok bahan baku outlet/gudang.
* `inventory:opname:manage` - Melakukan input penyesuaian stok opname.
* `inventory:transfer:request` - Mengajukan permintaan transfer stok antar-cabang/ke gudang pusat.
* `inventory:transfer:approve` - Menyetujui pengiriman mutasi stok.
* `inventory:po:create` - Membuat Purchase Order bahan ke supplier.
* `inventory:po:approve` - Otorisasi dan persetujuan PO pembelian (HQ).

### F. Modul Finance & Reports (`finance`)
* `finance:report:sales` - Laporan ringkasan penjualan harian/bulanan.
* `finance:report:pnl` - Laporan laba rugi konsolidasi / HPP operasional.
* `finance:payment_method:manage` - Mengatur rekening bank, gateway QRIS, dan EDC.

---

## 5. Strategi Performa & Caching di Golang

Menyimpan puluhan permission string di dalam token JWT akan membuat header HTTP membengkak menjadi 2KB–4KB di setiap request.

### Solusi Desain:
1. **JWT Payload Ringkas**:
   * Token JWT hanya menyimpan:
     ```json
     {
       "sub": "user-uuid",
       "role": "BRANCH_MANAGER",
       "role_id": "role-uuid",
       "organization_id": "org-uuid",
       "brand_id": "brand-uuid",
       "outlet_id": "outlet-uuid"
     }
     ```
2. **In-Memory Cache Layer (`sync.RWMutex` / Go Map / Redis)**:
   * Backend menyimpan struktur data: `map[string]map[string]bool` (`role_id -> set of permission codes`).
   * **Waktu Pengecekan**: `< 1 mikrodetik` (O(1) lookup di memory, 0 query ke database).
   * **Cache Invalidation**:
     * Saat service backend pertama kali bootstrap, seluruh permissions per role di-load ke memory.
     * Jika ada Owner yang mengubah permission di Backoffice, trigger event invalidasi (atau pub/sub Redis di multi-instance) untuk memuat ulang permission role tersebut.

---

## 6. Fiber v3 Middleware Design

### A. Middleware `RequirePermission`
```go
// internal/middleware/rbac.go
func RequirePermission(rbacService RBACService, requiredPerm string) fiber.Handler {
    return func(c fiber.Ctx) error {
        roleID := contextutil.GetUserRoleID(c)
        if roleID == "" {
            return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN", "peran pengguna tidak ditemukan")
        }

        if !rbacService.HasPermission(roleID, requiredPerm) {
            return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN_PERMISSION", "anda tidak memiliki izin untuk melakukan aksi ini")
        }

        return c.Next()
    }
}
```

### B. Middleware `SupervisorOverrideGuard` (FnB POS Void/Refund)
```go
// Memeriksa apakah user memiliki permission langsung, 
// ATAU jika tidak punya, memvalidasi PIN supervisor dari header X-Supervisor-PIN
func RequirePermissionOrSupervisorPIN(rbacService RBACService, perm string) fiber.Handler
```

---

## 7. Integrasi dengan Lookup Endpoints (Comboboxes)

Mengacu pada aturan di [rbac-lookup.md](../../.agents/rules/rbac-lookup.md):

1. **Endpoint `GET /api/v1/roles/lookup`**:
   * Digunakan oleh form penambahan staf di Backoffice.
   * Hanya mengembalikan: `id`, `name`, `code`, `tier`.
   * Memfilter `organization_id` yang sesuai dan `status = 'ACTIVE'`.
   * Bebas dari data sensitif.
2. **Endpoint `GET /api/v1/permissions/lookup`**:
   * Digunakan oleh halaman pengaturan custom role di Backoffice.
   * Mengembalikan daftar permission dikelompokkan berdasarkan `group` untuk kemudahan render checkbox matrix di frontend.

---

## 8. Roadmap Implementasi (Step-by-Step)

1. **Tahap 1: Migrasi Database (DDL)**
   * Buat migrasi SQL up/down:
     * `000005_create_rbac_tables.up.sql` (`permissions`, `roles`, `role_permissions`, `user_outlets`).
     * Alter table `users` menambahkan `role_id`.
2. **Tahap 2: Master Seeding**
   * Masukkan seluruh daftar permission awal.
   * Masukkan 8 System Roles (`OWNER`, `FINANCE_HQ`, `CENTRAL_PURCHASING`, `BRAND_MANAGER`, `BRANCH_MANAGER`, `CASHIER`, `BARISTA`, `WAITER`) beserta default permissions-nya.
   * Update seed data pengguna di `cmd/seed/main.go` agar mengaitkan `role_id` yang sesuai.
3. **Tahap 3: Modul RBAC (Clean Architecture)**
   * `internal/modules/rbac/domain`: Entity `Role`, `Permission`, interface repository.
   * `internal/modules/rbac/infrastructure`: GORM persistence & In-Memory RBAC Cache.
   * `internal/modules/rbac/usecase`: DTO, caching synchronization, role CRUD.
   * `internal/modules/rbac/delivery`: HTTP Handlers (`GET /roles`, `GET /roles/lookup`, `POST /roles`, `GET /permissions/lookup`).
4. **Tahap 4: Middleware Integration & Security Guard**
   * Buat `internal/middleware/rbac.go` (`RequirePermission`, `RequireAnyPermission`, `RequireSupervisorPIN`).
   * Pasang pada route yang relevan.
5. **Tahap 5: Testing & Dokumentasi**
   * Unit test usecase & caching layer.
   * Delivery test endpoint dengan Fiber test framework.
   * Update OpenAPI / Swagger specs di `docs/specs/backoffice.yaml`.
