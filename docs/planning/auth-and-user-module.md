# Planning & Brainstorming: Modul Auth & User

Dokumen ini merangkum hasil brainstorming, arsitektur, dan rencana implementasi untuk modul **Authentication** (`auth`) dan **User Management** (`user`) pada Hanjo ERP (Multi-Tenant FnB ERP & POS Backend).

---

## 1. Arsitektur & Pemisahan Modul

Berdasarkan Clean Architecture & Domain-Driven Design (DDD):
* **`internal/modules/user`**:
  * Mengelola lifecycle entity User/Staff.
  * CRUD karyawan/staf, update profile, ubah status aktif/nonaktif, pencarian staf berdasarkan organisasi dan outlet.
  * Domain Entity: `User`
  * Repository: `UserRepository` (GetByID, GetByEmail, List, Create, Update, Delete)
* **`internal/modules/auth`**:
  * Mengelola workflow autentikasi & otorisasi.
  * Register (Owner Onboarding), Login (Email + Password), POS Login (Quick PIN), Refresh Token (HttpOnly Cookie), Logout, `/me` profile check.
  * Dependency: Meng-inject `userdomain.UserRepository` (inversion of control, no direct DB/GORM cross-access).
  * Enkripsi: Argon2id via `pkg/password`.
  * Tokenizer: RS256 Asymmetric JWT via `pkg/jwt`.
  * Brute Force Throttling: `pkg/limiter`.

---

## 2. Multi-Tenancy & Onboarding Flow (3-Tier Model)

Mengacu pada arsitektur resmi di [docs/knowledge/multi-tenant-hierarchy.md](../knowledge/multi-tenant-hierarchy.md):
* **Hirarki Multi-Tenant**:
  1. `Organization` (Holding / Corporate Management - Level 1)
  2. `Brand` (Merek / Konsep Coffee Shop - Level 2)
  3. `Outlet` (Cabang Fisik / Lokasi Toko - Level 3)
  4. `User` (Pengguna sistem dengan scope tertentu)
* **Alur Registrasi (Self-Service Onboarding)**:
  * Endpoint: `POST /api/v1/auth/register`
  * Payload: `name`, `email`, `password`, `organization_name`, `phone` (opsional).
  * Mekanisme: Dijalankan dalam database transaction (`pkg/trx`):
    1. Membuat record baru di tabel `organizations`.
    2. Membuat record user baru di tabel `users` dengan role `OWNER`, terikat pada `organization_id` yang baru dibuat.
    3. Return access token & set HttpOnly refresh token cookie (otomatis login setelah register).
* **Alur Penambahan Staff (Back-Office / POS)**:
  * Dilakukan oleh `OWNER`, `BRAND_MANAGER`, atau `BRANCH_MANAGER` melalui `POST /api/v1/users`.
  * User baru otomatis terikat dengan `organization_id` pembuatnya, serta opsional `brand_id` dan `outlet_id`.

---

## 3. Role & Hak Akses (F&B / POS Context)

| Role | Scope | Cakupan Akses & Deskripsi |
| :--- | :--- | :--- |
| `OWNER` | Organization | Pemilik grup usaha; akses penuh konfigurasi grup, laporan finansial gabungan, seluruh brand, dan seluruh outlet. |
| `FINANCE_HQ` | Organization | Manajemen keuangan grup; purchase order pusat, vendor master, dan mutasi kas grup. |
| `BRAND_MANAGER`| Brand | Manajer brand spesifik; mengelola katalog menu master, resep, dan promo cabang di bawah brand tersebut. |
| `BRANCH_MANAGER`| Outlet | Manajer cabang fisik; mengelola operasional, stok cabang, kasir, dan laporan shift cabang. |
| `CASHIER` | Outlet | Kasir POS; mengoperasikan transaksi kasir, buka/tutup shift kasir, dan cetak struk di cabang fisik. |
| `BARISTA` / `KITCHEN` | Outlet | Bar/Dapur; melihat antrean pesanan via Kitchen Display System (KDS) dan mengubah status hidangan. |
| `WAITER` | Outlet | Pelayan; mencatat pesanan tamu di meja dan mengirim pesanan ke dapur/bar. |

---

## 4. Metode Autentikasi

1. **Email + Password (Web Dashboard & Admin)**:
   * Password di-hash menggunakan **Bcrypt** (`pkg/password`), hemat RAM dan ramah VPS kecil (migrasi ke Argon2id disiapkan pada roadmap scale-up).
   * Validasi password ketat via `pkg/validator` (minimal 8 karakter, huruf besar, huruf kecil, angka, simbol).
2. **Quick POS PIN (Kasir / Barista Fast Switching)**:
   * Pada mesin POS tablet, pergantian kasir/barista antar-shift membutuhkan login cepat tanpa mengetik email & password panjang.
   * Kolom `pin_hash` (4–6 digit angka) disiapkan di tabel `users`.
   * Endpoint POS login: `POST /api/v1/auth/pos/login` (menerima `outlet_id` + `user_id` + `pin`).

---

## 5. Token & Session Management

* **Access Token**:
  * RS256 JWT, masa berlaku 15 menit s/d 8 jam (sesuai konfigurasi).
  * Disimpan di client memory (Authorization Header: `Bearer <token>`).
  * Payload claims: `user_id`, `email`, `organization_id`, `brand_id`, `outlet_id`.
* **Refresh Token**:
  * JWT Refresh Token ditandatangani secara aman dengan secret/key, dikirim via **HttpOnly, Secure, SameSite=Lax Cookie** (`refresh_token`).
  * Masa berlaku: 7 hari.
* **Pencatatan Masa Depan (Redis Roadmap)**:
  * Redis dicatat dalam roadmap fase berikutnya untuk *distributed session store*, *cross-server rate limiting*, dan *caching*.
  * Untuk fase sekarang, sistem berjalan mandiri (*zero external dependency*) menggunakan PostgreSQL dan JWT + in-memory revocation.
* **Brute-Force & Rate Limiting**:
  * Menggunakan `pkg/limiter` (in-memory rate limiter) untuk membatasi percobaan login (misal 5 kali gagal dalam 15 menit per IP/akun).

---

## 6. Rancangan Skema Database (PostgreSQL Migrations)

Database PostgreSQL difokuskan murni untuk data transaksional dan master data. Migrasi di-split menjadi 4 file:

### 6.1 `000001_create_organizations_table` (Tier 1: Holding / Management)
```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;
```

### 6.2 `000002_create_brands_table` (Tier 2: Coffee Shop Brand / Konsep)
```sql
CREATE TABLE brands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    logo_url TEXT,
    tagline VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_brands_org_slug UNIQUE(organization_id, slug)
);

CREATE INDEX idx_brands_org_id ON brands(organization_id) WHERE deleted_at IS NULL;
```

### 6.3 `000003_create_outlets_table` (Tier 3: Cabang Fisik / Store)
```sql
CREATE TABLE outlets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    brand_id UUID NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    address TEXT NOT NULL,
    phone VARCHAR(50),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_outlets_org_code UNIQUE(organization_id, code)
);

CREATE INDEX idx_outlets_org_brand ON outlets(organization_id, brand_id) WHERE deleted_at IS NULL;
```

### 6.4 `000004_create_users_table` (Identity & Staff)
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    brand_id UUID REFERENCES brands(id) ON DELETE SET NULL,
    outlet_id UUID REFERENCES outlets(id) ON DELETE SET NULL,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    pin VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, INACTIVE, SUSPENDED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

-- Partial Unique Indexes per Organization
CREATE UNIQUE INDEX idx_users_org_email ON users(organization_id, email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_users_org_username ON users(organization_id, username) WHERE deleted_at IS NULL;

-- Hierarchy, Scope, and Lookup Indexes
CREATE INDEX idx_users_org_id ON users(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_brand_id ON users(brand_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_outlet_id ON users(outlet_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
```

---

## 7. Rencana Kontrak Endpoint API

### 7.1 Modul Auth (`/api/v1/auth`)
* `POST /api/v1/auth/login` (Public) - Login email & password (Dashboard Web untuk Owner, Manager, Finance).
* `POST /api/v1/auth/pos/login` (Public / Scoped) - Login kasir/barista via PIN di outlet spesifik.
* `POST /api/v1/auth/refresh` (Public via Cookie) - Rotasi & refresh access token baru via Redis.
* `POST /api/v1/auth/logout` (Protected) - Revoke refresh token di Redis & bersihkan cookie.
* `GET /api/v1/auth/me` (Protected) - Mendapatkan profile & hierarki context user yang sedang login.

> **Catatan Onboarding**: Sistem menerapkan model *Managed B2B / High-Touch ERP*. Tidak disediakan registrasi publik (`register`) untuk mencegah spam. Organisasi, Brand, Outlet, dan akun Owner awal di-provisioning melalui Database Seeder / Super-Admin CLI.

### 7.2 Modul User (`/api/v1/users`)
* `GET /api/v1/users` (Protected: `OWNER`, `BRANCH_MANAGER`) - List staf per organisasi/outlet dengan pagination & filter.
* `POST /api/v1/users` (Protected: `OWNER`, `BRANCH_MANAGER`) - Tambah staf/kasir/barista baru beserta PIN POS.
* `GET /api/v1/users/:id` (Protected: `OWNER`, `BRANCH_MANAGER`) - Detail informasi staf.
* `PUT /api/v1/users/:id` (Protected: `OWNER`, `BRANCH_MANAGER`) - Update data staf (nama, phone, role, outlet, PIN).
* `DELETE /api/v1/users/:id` (Protected: `OWNER`) - Soft-delete staf.
* `PATCH /api/v1/users/:id/status` (Protected: `OWNER`, `BRANCH_MANAGER`) - Update status (`ACTIVE`/`INACTIVE`) + force logout sesi di Redis.

---

## 8. Database Seeder & Akun Demo

Seeder CLI dapat dijalankan melalui perintah `make seed` (`cmd/seed/main.go`).

| Tipe Akun | Email / Username | Password / PIN | Scope | Deskripsi |
| :--- | :--- | :--- | :--- | :--- |
| **Developer / Super Admin** | `dev@hanjo.id` (`developer`) | `Password123!` | Global (`brand_id: nil`, `outlet_id: nil`) | Akses penuh superadmin / developer untuk maintenance & bypass |
| **Owner (Holding HQ)** | `owner@harmoni.id` (`owner`) | `Password123!` | Organization HQ | Akses owner holding multi-brand & multi-outlet |
| **Brand Manager** | `manager.daily@harmoni.id` | `Password123!` | Brand (`Daily Coffee`) | Akses manajer brand |
| **Branch Manager** | `manager.kemang@harmoni.id` | `Password123!` | Outlet (`Daily Coffee - Kemang`) | Akses kepala cabang outlet |
| **Kasir POS** | `kasir.kemang@harmoni.id` (`kasir.kemang`) | PIN: `123456` / `Password123!` | Outlet (`Daily Coffee - Kemang`) | Akses kasir POS tablet |
| **Barista / Kitchen** | `barista.kemang@harmoni.id` (`barista.kemang`) | PIN: `654321` / `Password123!` | Outlet (`Daily Coffee - Kemang`) | Akses kitchen / bar display |
