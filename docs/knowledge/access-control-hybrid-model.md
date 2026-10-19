# Access Control Architecture Knowledge: Hybrid Model (Scoped RBAC + Contextual Policy Guards)

Dokumen ini adalah **Global Knowledge** resmi untuk arsitektur pengontrolan akses (*Access Control*) pada Hanjo ERP & POS Backend. Dokumen ini mendefinisikan keputusan arsitektural pemilihan model **Hybrid (Scoped Dynamic RBAC + Contextual Policy Guards / ABAC)**, membedah 3-layer evaluasi otorisasi, serta menjadi panduan implementasi bagi developer dan AI agent di masa mendatang.

---

## 1. Latar Belakang & Keputusan Arsitektur

Dalam sistem enterprise multi-tenant Food & Beverage (FnB) ERP dan Point of Sale (POS), sistem otorisasi menghadapi tantangan unik:
1. **Multi-Cabang & Multi-Brand**: Kasir di Cabang Kemang tidak boleh melihat atau memodifikasi transaksi di Cabang Sudirman, meskipun jabatan mereka sama-sama *"Kasir"*.
2. **Kebutuhan Frontline Cepat**: POS membutuhkan respon otorisasi instan (< 1 milidetik) agar tidak menghambat antrean kasir.
3. **User-Friendly Backoffice**: Owner restoran bukan teknisi perangkat lunak; mereka membutuhkan antarmuka pengelolaan peran yang sederhana (membuat nama peran dan mencentang izin akses dari tabel/matriks).
4. **Otorisasi Lapangan Kondisional**: Operasional kasir membutuhkan otorisasi fleksibel di tempat (*Supervisor PIN Override*) saat aksi terlarang seperti *Void* atau *Manual Discount* dilakukan.

### Mengapa Bukan Pure RBAC?
Pada **Pure RBAC** (*Role-Based Access Control*), hak akses diikat murni pada nama peran (`User -> Role -> Permission`).
* **Masalah**: Terjadi **Role Explosion**. Jika batasan cabang dimasukkan ke dalam peran, sistem harus membuat ribuan peran spesifik: `CASHIER_KEMANG`, `CASHIER_SENOPATI`, `BRAND_MGR_SENJA`, dst. Hal ini membuat database membengkak dan manajemen izin menjadi tidak terkendali saat cabang bertambah.
* **Kelemahan Lain**: Pure RBAC tidak dapat mengevaluasi kondisi dinamis (misal: apakah shift kasir berstatus OPEN, atau apakah nominal void melebihi limit tertentu).

### Mengapa Bukan Pure PBAC / ABAC?
Pada **Pure PBAC / ABAC** (*Policy-Based / Attribute-Based Access Control*), setiap aksi dievaluasi menggunakan *rule engine* dinamis (seperti OPA/Rego atau AWS Cedar) berdasarkan atribut subjek, objek, aksi, dan lingkungan.
* **Masalah**: Terlalu rumit untuk dipahami oleh pemilik bisnis FnB dan pengelola restoran. Sangat sulit menyediakan UI yang mudah dipakai jika setiap hak akses berupa script/aturan logika kondisional yang rumit.
* **Overhead Performa**: Evaluasi mesin kebijakan dinamis untuk setiap request HTTP menambah latensi yang tidak diinginkan pada operasional kasir.

### Solusi: HYBRID (Scoped RBAC + Contextual Policy Guards)
Hanjo ERP memilih pendekatan **Hybrid** yang menggabungkan keunggulan keduanya:
* **Komponen RBAC**: Mengelola kapabilitas teknis (*"What can you do?"*) melalui relasi Peran & Izin atomik. Sangat mudah dipetakan ke UI Backoffice (centang izin) dan di-cache dalam memori dengan kecepatan $O(1)$ (< 1 mikrodetik).
* **Komponen ABAC & Scoping**: Mengelola wilayah data (*"Where can you do it?"*) melalui atribut hierarki 3-Tier (`organization_id`, `brand_id`, `outlet_id`) yang diinjeksikan secara otomatis ke filter query database.
* **Komponen Contextual Policy Guards**: Mengelola aturan situasional (*"When & Under what conditions?"*) seperti validasi PIN Supervisor di tempat untuk aksi terbatas (Void/Refund).

---

## 2. Tiga Layer Evaluasi Otorisasi (The 3-Tier Security Layers)

Setiap request HTTP yang masuk ke protected endpoint diproses melalui 3 layer evaluasi berurutan:

```text
HTTP Request Masuk (Client POS / Backoffice)
     │
     ▼
┌────────────────────────────────────────────────────────┐
│ Layer 1: RBAC Capability Check ("WHAT")                │
│ - Apakah role user memiliki kode permission ini?       │
│ - Evaluasi: In-Memory Go Hashmap O(1) (< 1 mikrodetik) │
│ - Contoh: HasPermission(roleID, "pos:order:void")      │
└───────────────────────────┬────────────────────────────┘
                            │ Lolos (atau masuk ke Supervisor PIN Fallback)
                            ▼
┌────────────────────────────────────────────────────────┐
│ Layer 2: Tenant Scoping ("WHERE")                      │
│ - Apakah data target berada di wilayah izin user?      │
│ - Evaluasi: GORM Query Scope Filter (DB Index)         │
│ - Tier 1 (Org)    : WHERE organization_id = ?          │
│ - Tier 2 (Brand)  : AND brand_id = ?                   │
│ - Tier 3 (Outlet) : AND outlet_id = ?                  │
└───────────────────────────┬────────────────────────────┘
                            │ Lolos
                            ▼
┌────────────────────────────────────────────────────────┐
│ Layer 3: Contextual Policy Guards ("WHEN & HOW")       │
│ - Evaluasi Aturan Kondisional & Status Bisnis:         │
│   1. Shift Status: Apakah shift kasir masih OPEN?      │
│   2. SPV Override: Jika Kasir tidak punya izin void,   │
│      apakah ada validasi PIN Supervisor dari SPV toko? │
│   3. Threshold: Apakah diskon melebihi persentase max? │
└───────────────────────────┬────────────────────────────┘
                            │ Lolos
                            ▼
                   Eksekusi Handler Usecase
```

---

## 3. Rincian Layer Otorisasi

### 3.1. Layer 1: RBAC Capability (Hak Akses Atomik)
* **Entitas**:
  * `permissions`: Daftar aksi atomik sistem dengan format standar `<module>:<resource>:<action>` (contoh: `pos:order:void`, `catalog:item:manage`, `inventory:stock:transfer`).
  * `roles`: Peran pengguna, baik bawaan sistem (*System Defaults*) maupun peran kustom yang dibuat Owner restoran.
  * `role_permissions`: Tabel penghubung (*junction table*) antara peran dan daftar izin.
* **Mekanisme Performa (In-Memory Cache)**:
  * Backend memuat seluruh pemetaan `role_id -> set of permissions` ke dalam memori aplikasi (`map[string]map[string]bool` dilindungi `sync.RWMutex`).
  * Pengecekan izin tidak pernah melakukan query SQL ke database di setiap request, sehingga latensi verifikasi < 1 mikrodetik.
  * Saat Owner mengubah izin peran di Backoffice, cache di-invalidasi dan di-reload secara otomatis.

### 3.2. Layer 2: Tenant Scoping (Wilayah Data 3-Tier)
Hierarki multi-tenant membagi ruang lingkup pengguna ke dalam 3 tier:

| Role Tier | Penugasan Pengguna | Contoh Peran | Cakupan Akses Data |
| :--- | :--- | :--- | :--- |
| **`ORGANIZATION`** | `organization_id` wajib<br>`brand_id = NULL`<br>`outlet_id = NULL` | `OWNER`<br>`FINANCE_HQ`<br>`CENTRAL_PURCHASING` | Mengakses seluruh data konsolidasi holding di semua brand dan outlet. |
| **`BRAND`** | `organization_id` wajib<br>`brand_id` wajib<br>`outlet_id = NULL` | `BRAND_MANAGER`<br>`RND_CHEF` | Mengakses data katalog menu, bahan baku, dan promo khusus brand yang ditugaskan di seluruh cabang brand tersebut. |
| **`OUTLET`** | `organization_id` wajib<br>`brand_id` wajib<br>`outlet_id` wajib | `BRANCH_MANAGER`<br>`CASHIER`<br>`BARISTA`<br>`WAITER` | Terkunci hanya pada 1 outlet fisik tempat pengguna ditugaskan (transaksi harian, kas register, meja, stok cabang). |

* **Multi-Outlet / Area Manager**:
  * Pengguna tingkat manajerial cabang yang membawahi lebih dari satu outlet (misal: Area Manager membawahi 3 cabang) dicatat pada tabel `user_outlets`.
  * Sistem mengizinkan pergantian konteks (*active outlet context*) sepanjang outlet tersebut terdaftar pada relasi `user_outlets`.

### 3.3. Layer 3: Contextual Policy Guards (Skenario Lapangan FnB)
Layer ini menangani kasus-kasus spesifik yang tidak dapat diselesaikan oleh centang izin biasa:

1. **Supervisor PIN Override (Kasus Void & Diskon Kasir)**:
   * Kasir garis depan (`CASHIER`) tidak memiliki izin `pos:order:void`.
   * Saat pelanggan membatalkan pesanan, kasir tidak perlu logout. POS mengirimkan request dengan header `X-Supervisor-PIN`.
   * Sistem memeriksa apakah PIN tersebut milik user dengan role berwenang (`BRANCH_MANAGER` / `SUPERVISOR`) di outlet yang sama.
   * Jika valid, aksi dijalankan dan rekam jejak dicatat otomatis ke tabel `supervisor_override_logs` (siapa kasir yang meminta, siapa supervisor yang menyetujui, alasan void, dan timestamp).
2. **Shift Readiness Guard**:
   * Aksi pembuatan pesanan kasir (`pos:order:create`) atau penerimaan pembayaran (`pos:order:payment`) mewajibkan kasir memiliki sesi kasir aktif (`shift_id` berstatus `OPEN`).
3. **Price/Discount Threshold Guard**:
   * Diskon manual di atas nominal atau persentase tertentu (misal: > 20%) otomatis mewajibkan otorisasi tambahan dari Store Manager.

---

## 4. Matriks Peran Default Sistem (8 System Roles)

Berikut adalah 8 peran standar yang disediakan secara *out-of-the-box* oleh Hanjo ERP:

| Role Code | Tier | Lingkup Akses Data | Ringkasan Tanggung Jawab & Izin Utama |
| :--- | :--- | :--- | :--- |
| **`OWNER`** | `ORGANIZATION` | Seluruh Organisasi | Hak akses penuh administratif, manajemen organisasi, keuangan konsolidasi, audit trail, pembuatan custom role. |
| **`FINANCE_HQ`** | `ORGANIZATION` | Seluruh Organisasi | Laporan laba rugi konsolidasi, arus kas, rekonsiliasi pembayaran payment gateway, penutupan buku akuntansi. |
| **`CENTRAL_PURCHASING`** | `ORGANIZATION` | Seluruh Organisasi | Manajemen supplier, Purchase Order skala holding, penerimaan bahan baku di Gudang Pusat (Central Warehouse). |
| **`BRAND_MANAGER`** | `BRAND` | Spesifik 1 Brand | Pengelolaan kategori menu, formula resep (BOM), penyesuaian harga brand, dan manajemen promo pemasaran brand. |
| **`BRANCH_MANAGER`** | `OUTLET` | Spesifik 1 Outlet | Pengelolaan operasional harian cabang, persetujuan opname stok lokal, otorisasi PIN supervisor (Void/Refund), jadwal shift karyawan cabang. |
| **`CASHIER`** | `OUTLET` | Spesifik 1 Outlet | Pembukaan/penutupan shift laci kasir, pencatatan transaksi POS, cetak bill, proses pembayaran (Tunai, QRIS, Kartu). |
| **`BARISTA` / `KITCHEN`** | `OUTLET` | Spesifik 1 Outlet | Tampilan pesanan dapur (Kitchen Display System / KDS), pembaruan status pengerjaan pesanan (*Cooking*, *Ready*). |
| **`WAITER`** | `OUTLET` | Spesifik 1 Outlet | Pemesanan meja (*Table Order*), pengecekan status ketersediaan meja, pengantaran pesanan ke tamu. |

---

## 5. Pemisahan Endpoint Administrasi vs Operational Lookup

Mengacu pada prinsip arsitektural di [.agents/rules/rbac-lookup.md](../../.agents/rules/rbac-lookup.md):

1. **Administrative Endpoints (`GET /api/v1/<entities>`)**:
   * Ditujukan untuk dashboard Backoffice HQ dan Store Manager.
   * Memuat data lengkap, sensitif (gaji, cost HPP, nomor telepon, status audit, deleted records).
   * Dilindungi guard ketat `RequirePermission`.
2. **Lookup Endpoints (`GET /api/v1/<entities>/lookup`)**:
   * Ditujukan untuk komponen *combobox/dropdown* pada UI POS dan Backoffice (misal: kasir memilih nama waiter pelayan meja, manajer memilih nama staf).
   * **Bebas Data Sensitif**: Hanya mengembalikan field publik (`id`, `name`, `code`, `tier`).
   * **Filter Otomatis**: Hanya mengembalikan record dengan `status = 'ACTIVE'`.
   * **Terisolasi Tenant**: Otomatis dibatasi oleh `organization_id` dan `outlet_id` pemanggil.
   * Dapat diakses oleh peran tingkat rendah (`CASHIER`, `WAITER`) tanpa memicu error `403 Forbidden`.

---

## 6. Panduan untuk Developer & AI Agent

Saat menambah modul atau endpoint baru di Hanjo ERP:

1. **Tentukan Taksonomi Permission Baru**:
   * Format penamaan wajib: `<module>:<resource>:<action>` (contoh: `table:layout:manage`, `inventory:waste:create`).
   * Masukkan ke dalam seeder master permissions.
2. **Pasang Middleware Sesuai Karakteristik Endpoint**:
   * Operasi administrasi biasa: gunakan `middleware.RequirePermission("module:resource:action")`.
   * Operasi POS sensitif yang mendukung override: gunakan `middleware.RequirePermissionOrSupervisorPIN("pos:order:void")`.
   * Endpoint dropdown/combobox: sediakan route `GET /<resource>/lookup` terpisah yang didaftarkan **sebelum** route `/:id`.
3. **Enforce Tenant Scoping pada Query**:
   * Selalu gunakan `pkg/contextutil` untuk mengambil `organization_id`, `brand_id`, dan `outlet_id` dari JWT Claims.
   * Pastikan query GORM menerapkan isolasi tenant secara konsisten menggunakan `pkg/gormscope`.
4. **Hindari Role Explosion**:
   * Jangan pernah membuat role terpisah hanya untuk membedakan cabang atau brand. Gunakan atribut tenant scoping untuk membatasi wilayah data.
