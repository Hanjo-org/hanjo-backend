# Multi-Tenant Hierarchy & Architecture Knowledge: 3-Tier FnB Model

Dokumen ini adalah **Global Knowledge** resmi untuk arsitektur multi-tenant pada Hanjo ERP & POS Backend. Dokumen ini mendefinisikan struktur organisasi, batasan hierarki, tanggung jawab tiap entitas, peran pengguna (RBAC), serta implikasinya terhadap database dan alur autentikasi.

---

## 1. Latar Belakang & Kebutuhan Bisnis (Real-World FnB Context)

Sistem ini dirancang khusus untuk mengakomodasi model bisnis **Multi-Brand / Multi-Concept under One Management** (Holding Management / Hospitality Group).

### Kasus Bisnis Client:
* Seorang pengusaha atau grup manajemen F&B memiliki **beberapa merek coffee shop / restoran yang berbeda nama dan konsep** dalam 1 payung manajemen terintegrasi.
* **Contoh Kasus**:
  * **Brand A**: *"Senja Coffee & Roastery"* (Spesialis manual brew & beans).
  * **Brand B**: *"Daily Coffee"* (Konsep es kopi susu & grab-and-go).
  * **Brand C**: *"Dua Sudut Cafe & Eatery"* (Kafe dine-in dengan hidangan berat).
* **Multi-Cabang pada Satu Brand**:
  * Brand B ("Daily Coffee") memiliki **banyak cabang fisik** di berbagai lokasi:
    * Cabang Kemang (Alamat A)
    * Cabang Sudirman (Alamat B)
    * Cabang BSD (Alamat C)
* **Kebutuhan Terintegrasi**:
  1. **Holding Finance**: Owner / manajemen pusat dapat melihat laporan konsolidasi total (omset gabungan Brand A + B + C), sekaligus membedah performa per brand dan per cabang.
  2. **Central Supply Chain**: Manajemen pusat dapat mengelola Gudang Pusat (Central Warehouse / Central Kitchen) untuk pengadaan bahan baku skala besar (biji kopi, susu, sirup, cup), lalu mentransfer stok ke masing-masing cabang.
  3. **Master Menu Efisien**: Menu Coffee Shop B cukup di-input 1 kali di level brand, dan otomatis tersedia di semua cabang Coffee Shop B tanpa perlu input manual di tiap outlet.
  4. **Identitas Mandiri di Struk**: Struk kasir di cabang Kemang akan mencetak nama Brand B ("Daily Coffee") beserta alamat fisik Cabang Kemang.

---

## 2. Struktur 3-Tier Hierarchy

Untuk menyelesaikan kebutuhan di atas secara terukur dan *scalable*, sistem menerapkan arsitektur **3-Tier Hierarchy**:

```text
[TIER 1: HOLDING / MANAGEMENT]
Organization ("PT Harmoni Rasa Nusantara")
│
├── Central Warehouse (Gudang Pusat / Raw Ingredients)
│
├── [TIER 2: BRAND / CONCEPT]
│   Brand A ("Senja Coffee & Roastery")
│   └── [TIER 3: OUTLET / STORE]
│       └── Outlet Senopati (Jl. Senopati No. 12)
│
├── [TIER 2: BRAND / CONCEPT]
│   Brand B ("Daily Coffee")
│   ├── [TIER 3: OUTLET / STORE]
│   │   └── Outlet Kemang (Jl. Kemang Raya No. 5)
│   ├── [TIER 3: OUTLET / STORE]
│   │   └── Outlet Sudirman (Jl. Jend. Sudirman Kav. 88)
│   └── [TIER 3: OUTLET / STORE]
│       └── Outlet BSD (Ruko Foresta BSD No. 10)
│
└── [TIER 2: BRAND / CONCEPT]
    Brand C ("Dua Sudut Eatery")
    └── [TIER 3: OUTLET / STORE]
        └── Outlet Tebet (Jl. Tebet Timur No. 3)
```

---

## 3. Matriks Tanggung Jawab Tiap Entitas

| Level Entitas | Cakupan Bisnis | Tanggung Jawab & Fitur Utama |
| :--- | :--- | :--- |
| **Tier 1: `Organization`**<br>*(Holding / Manajemen)* | Enterprise Group / Tenant | • Konsolidasi laporan laba-rugi & omset seluruh brand.<br>• Master akun supplier / vendor terpusat.<br>• Central procurement & gudang pusat bahan baku.<br>• Manajemen lisensi / subscription sistem.<br>• Manajemen role global (Owner, CFO, Central Procurement). |
| **Tier 2: `Brand`**<br>*(Merek / Unit Bisnis)* | Konsep Kafe / Merek Dagang | • Master Katalog Menu & Resep (BOM - Bill of Materials).<br>• Kategori Produk (Coffee, Non-Coffee, Pastry).<br>• Identitas visual: Logo brand, warna tema struk kasir, tagline.<br>• Program promosi / diskon spesifik brand. |
| **Tier 3: `Outlet`**<br>*(Cabang Fisik / Toko)* | Lokasi Operasional / POS | • Alamat fisik toko, kontak, koordinat GPS.<br>• Perangkat kasir (POS Terminal / Register).<br>• Manajemen meja (Table Layout, QR Table Ordering).<br>• Stok harian / *local inventory* (sisa susu, biji kopi harian).<br>• Buka/tutup shift kasir & rekap kas (*cash float*).<br>• Cetak struk belanja pelanggan. |

---

## 4. Role-Based Access Control (RBAC) & Scope Mapping

Setiap user di dalam sistem diikat oleh tingkat otoritas (*scope*) yang jelas:

```text
User Scope Hierarchy:
  └── Organization Level (Owner, Finance HQ, Central Purchasing)
        └── Brand Level (Brand Manager, Head of Coffee/Menu)
              └── Outlet Level (Branch Manager, Cashier, Barista, Kitchen, Waiter)
```

### Rincian Peran Pengguna:

1. **`OWNER` (Scope: Organization)**
   * Memiliki akses tak terbatas ke seluruh organisasi, seluruh brand, dan seluruh outlet.
   * Melihat analitik grup, laporan finansial gabungan, audit log, dan konfigurasi master.

2. **`FINANCE_HQ` / `CENTRAL_PURCHASING` (Scope: Organization)**
   * Mengelola transaksi pembelian bahan baku grup, utang/piutang supplier, dan persetujuan purchase order (PO) gudang pusat.

3. **`BRAND_MANAGER` (Scope: Brand)**
   * Bertanggung jawab atas 1 brand tertentu (misal Brand B).
   * Mengatur resep menu, harga dasar produk, promo brand, dan memantau performa cabang-cabang di bawah brand tersebut. Tidak dapat melihat data Brand A atau Brand C.

4. **`BRANCH_MANAGER` (Scope: Outlet)**
   * Mengelola 1 cabang fisik (misal Brand B - Cabang Kemang).
   * Mengatur jadwal staf outlet, permintaan transfer stok (*stock request*) ke gudang pusat, serta review laporan shift kasir harian.

5. **`CASHIER` (Scope: Outlet)**
   * Mengoperasikan POS tablet di cabang fisik bersangkutan.
   * Buka kasir, melayani pembayaran tunai/QRIS/kartu, cetak struk, tutup shift kasir.

6. **`BARISTA` / `KITCHEN` (Scope: Outlet)**
   * Mengakses Kitchen Display System (KDS) / Bar Display di cabang fisik bersangkutan.
   * Mengubah status pesanan (*Queue ➔ Preparing ➔ Ready to Serve*).

7. **`WAITER` (Scope: Outlet)**
   * Menginput pesanan meja tamu via tablet/smartphone di cabang fisik bersangkutan.

---

## 5. Rancangan Skema Database (PostgreSQL)

Setiap tabel mengimplementasikan standar Clean Architecture Hanjo ERP: UUID primary key, audit timestamps (`created_at`, `updated_at`, `deleted_at`), dan indeks teroptimasi.

### 5.1 Tabel `organizations` (Tier 1)
```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, SUSPENDED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;
```

### 5.2 Tabel `brands` (Tier 2)
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

### 5.3 Tabel `outlets` (Tier 3)
```sql
CREATE TABLE outlets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    brand_id UUID NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL, -- Misal: "DC-KMG" untuk Daily Coffee Kemang
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

### 5.4 Tabel `users` (Multi-Scope Identity)
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    brand_id UUID REFERENCES brands(id) ON DELETE SET NULL,     -- Nullable: terisi jika Brand Manager
    outlet_id UUID REFERENCES outlets(id) ON DELETE SET NULL,   -- Nullable: terisi jika Staf Outlet
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    pin VARCHAR(255),                                           -- 4-6 digit angka untuk POS quick switch
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
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

## 6. Implikasi terhadap Arsitektur Auth & Token Claims

1. **JWT Payload Claims (`pkg/jwt`)**:
   * Token JWT RS256 harus memuat konteks hierarki lengkap:
     ```json
     {
       "sub": "user-uuid",
       "user_id": "user-uuid",
       "email": "kasir.kemang@dailycoffee.id",
       "role": "CASHIER",
       "organization_id": "org-uuid",
       "brand_id": "brand-uuid",
       "outlet_id": "outlet-uuid",
       "exp": 1776432000
     }
     ```
2. **Context Propagation (`pkg/contextutil`)**:
   * Helper harus menyediakan akses cepat ke:
     * `contextutil.GetOrganizationID(c)`
     * `contextutil.GetBrandID(c)`
     * `contextutil.GetOutletID(c)`
3. **Switch Outlet / Switch Brand Context**:
   * Pengguna level `OWNER` memiliki akses ke semua brand dan outlet. Ketika Owner membuka dashboard untuk menginspeksi outlet tertentu, frontend dapat mengirimkan header `X-Outlet-ID` atau memanggil endpoint `POST /api/v1/auth/switch-context` untuk memperoleh token scoped sementara.
4. **Session & Token Storage (Redis)**:
   * Refresh token & sesi login aktif dikelola via **Redis** (`auth:refresh:<token_hash>`) dengan TTL native (7 hari).
   * Token revocation & force logout dikelola via set key `auth:user_sessions:<user_id>`, sehingga tidak memerlukan tabel `user_tokens` di PostgreSQL.

---

## 7. Rangkuman Prinsip untuk Pengembangan Fitur Masa Depan

* **Produk & Menu (Catalog Module)**:
  * Entitas `Product` berelasi langsung dengan `brand_id`.
  * Cabang (`Outlet`) tidak membuat master menu baru, melainkan memiliki tabel pivot `outlet_products` untuk mengatur *availability* (stok habis/aktif) dan *custom price override* (jika harga di bandara berbeda dengan ruko).
* **Inventory & Purchasing**:
  * Bahan baku master (`Item`) berada di level `Organization`.
  * Setiap `Outlet` dan `Central Warehouse` bertindak sebagai lokasi penyimpanan (*Storage Location / Stock Warehouse*) yang menerima mutasi stok via *Stock Transfer*.
* **Pesanan & Transaksi (Order / POS Module)**:
  * Setiap transaksi `Order` wajib merekam `organization_id`, `brand_id`, dan `outlet_id` agar data analitik dapat diagregasi secara instan di level mana pun.
