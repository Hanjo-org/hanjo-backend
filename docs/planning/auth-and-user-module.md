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

## 2. Multi-Tenancy & Onboarding Flow

* **Hirarki Multi-Tenant**:
  1. `Organization` (Tenant / Brand FnB level)
  2. `Outlet` (Cabang / Toko fisik)
  3. `User` (Karyawan / Pengguna sistem)
* **Alur Registrasi (Self-Service Onboarding)**:
  * Endpoint: `POST /api/v1/auth/register`
  * Payload: `name`, `email`, `password`, `organization_name`, `phone` (opsional).
  * Mekanisme: Dijalankan dalam database transaction (`pkg/trx`):
    1. Membuat record baru di tabel `organizations`.
    2. Membuat record user baru di tabel `users` dengan role `OWNER`, terikat pada `organization_id` yang baru dibuat.
    3. Return access token & set HttpOnly refresh token cookie (otomatis login setelah register).
* **Alur Penambahan Staff (Back-Office / POS)**:
  * Dilakukan oleh `OWNER` atau `MANAGER` melalui `POST /api/v1/users`.
  * User baru otomatis terikat dengan `organization_id` pembuatnya (dan opsional `outlet_id`).

---

## 3. Role & Hak Akses (F&B / POS Context)

| Role | Cakupan Akses | Deskripsi |
| :--- | :--- | :--- |
| `OWNER` | Multi-Outlet / Enterprise | Pemilik brand/usaha; akses penuh konfigurasi, laporan finansial, outlet, dan seluruh staf. |
| `MANAGER` | Single / Multi-Outlet | Manajer cabang; mengelola stok, kasir, pesanan, dan laporan operasional di cabangnya. |
| `CASHIER` | Single Outlet | Kasir POS; mengoperasikan transaksi meja/takeaway, buka/tutup kasir (shift), cetak struk. |
| `KITCHEN` | Single Outlet | Dapur; melihat antrean pesanan via Kitchen Display System (KDS) dan mengubah status hidangan. |
| `WAITER` | Single Outlet | Pelayan; mencatat pesanan tamu di meja, melihat status hidangan, mengirim pesanan ke dapur. |

---

## 4. Metode Autentikasi

1. **Email + Password (Web Dashboard & Admin)**:
   * Password di-hash menggunakan **Argon2id** (`pkg/password`).
   * Validasi password ketat via `pkg/validator` (minimal 8 karakter, huruf besar, huruf kecil, angka, simbol).
2. **Quick POS PIN (Kasir / Waiter Fast Switching)**:
   * Pada mesin POS tablet, pergantian kasir/waiter antar-shift membutuhkan login cepat tanpa mengetik email & password panjang.
   * Kolom `pin_hash` (4–6 digit angka) disiapkan di tabel `users`.
   * Endpoint POS login: `POST /api/v1/auth/pos/login` (menerima `outlet_id` + `pin`).

---

## 5. Token & Session Management

Sesuai aturan keamanan pada `.agents/rules/authentication.md`:
* **Access Token**:
  * RS256 JWT, masa berlaku 15 menit s/d 1 jam.
  * Disimpan di client memory (Authorization Header: `Bearer <token>`).
  * Payload claims: `user_id`, `email`, `role`, `organization_id`, `outlet_id`.
* **Refresh Token**:
  * String acak berkriptografi tinggi (crypto/rand).
  * Disimpan di database (`user_tokens`) dalam bentuk **SHA-256 Hash**.
  * Dikirim dan diterima via **HttpOnly, Secure, SameSite=Lax Cookie** (`refresh_token`).
  * Masa berlaku: 7 hari.
  * Fitur rotasi: Setiap kali refresh token digunakan, token lama di-revoke dan digantikan token baru untuk mencegah replay attack.
* **Brute-Force Protection**:
  * Menggunakan `pkg/limiter` untuk membatasi percobaan login (misal 5 kali gagal dalam 15 menit per IP/akun).

---

## 6. Rancangan Skema Database (PostgreSQL Migration)

### 6.1 Tabel `organizations`
```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;
```

### 6.2 Tabel `users`
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    outlet_id UUID, -- nullable untuk Owner/HQ staff
    name VARCHAR(150) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    pin_hash VARCHAR(255),
    phone VARCHAR(50),
    role VARCHAR(50) NOT NULL DEFAULT 'CASHIER',
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, INACTIVE, SUSPENDED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_users_org_email UNIQUE(organization_id, email)
);

CREATE INDEX idx_users_org_role ON users(organization_id, role) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
```

### 6.3 Tabel `user_tokens`
```sql
CREATE TABLE user_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address VARCHAR(45),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX idx_user_tokens_hash ON user_tokens(token_hash);
CREATE INDEX idx_user_tokens_expires ON user_tokens(expires_at);
```

---

## 7. Rencana Kontrak Endpoint API

### 7.1 Modul Auth (`/api/v1/auth`)
* `POST /api/v1/auth/register` (Public) - Register Owner & Organisasi baru.
* `POST /api/v1/auth/login` (Public) - Login email & password.
* `POST /api/v1/auth/pos/login` (Public / Outlet Scoped) - Login kasir/waiter via PIN.
* `POST /api/v1/auth/refresh` (Public via Cookie) - Refresh access token baru.
* `POST /api/v1/auth/logout` (Protected) - Revoke refresh token & bersihkan cookie.
* `GET /api/v1/auth/me` (Protected) - Mendapatkan profile & hak akses user yang sedang login.

### 7.2 Modul User (`/api/v1/users`)
* `GET /api/v1/users` (Protected: `OWNER`, `MANAGER`) - List staf per organisasi/outlet dengan pagination & filter.
* `POST /api/v1/users` (Protected: `OWNER`, `MANAGER`) - Tambah staf/kasir/waiter baru.
* `GET /api/v1/users/:id` (Protected: `OWNER`, `MANAGER`) - Detail informasi staf.
* `PUT /api/v1/users/:id` (Protected: `OWNER`, `MANAGER`) - Update data staf (nama, phone, role, outlet).
* `DELETE /api/v1/users/:id` (Protected: `OWNER`) - Soft-delete staf.
* `PATCH /api/v1/users/:id/status` (Protected: `OWNER`, `MANAGER`) - Update status (`ACTIVE`/`INACTIVE`).
