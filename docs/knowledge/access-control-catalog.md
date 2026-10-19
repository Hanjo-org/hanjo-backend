# Access Control Catalog & Permission Matrix Knowledge

Dokumen ini adalah **Katalog Master Access Control** resmi untuk Hanjo ERP & POS Backend. Dokumen ini mendokumentasikan seluruh daftar kode izin (*Permissions*), 9 peran sistem (*System Roles*), serta matriks pemetaan izin (*Role-Permission Matrix*) yang dijadikan acuan baku dalam database seeder, pengecekan otorisasi (*middleware guard*), dan penugasan peran pengguna.

---

## 1. Katalog Master Permissions (26 Atomic Capabilities)

Setiap izin didefinisikan secara atomik mengikuti konvensi: `<module>:<resource>:<action>`

| No | Module / Group | Permission Code | Label UI | Deskripsi |
| :---: | :--- | :--- | :--- | :--- |
| 1 | `user` | `user:staff:read` | Lihat Daftar Karyawan | Melihat data karyawan, profil, dan penugasan cabang. |
| 2 | `user` | `user:staff:create` | Tambah Karyawan Baru | Mendaftarkan pengguna baru dan mengaitkan role serta cabang. |
| 3 | `user` | `user:staff:update` | Ubah Data Karyawan | Memperbarui profil staf, mengubah role, atau mutasi cabang. |
| 4 | `user` | `user:staff:delete` | Nonaktifkan Karyawan | Menandai staf nonaktif atau menghapus akun (*soft delete*). |
| 5 | `rbac` | `rbac:role:read` | Lihat Daftar Peran & Izin | Melihat daftar role dan permission matrix di Backoffice. |
| 6 | `rbac` | `rbac:role:manage` | Kelola Custom Role | Membuat, mengubah, dan menghapus custom role organisasi. |
| 7 | `org` | `org:brand:read` | Lihat Daftar Brand | Melihat profil brand, logo, dan identitas konsep FnB. |
| 8 | `org` | `org:brand:manage` | Kelola Brand | Menambah atau memperbarui profil dan konfigurasi brand. |
| 9 | `org` | `org:outlet:read` | Lihat Daftar Outlet | Melihat daftar cabang fisik, alamat, dan nomor telepon. |
| 10 | `org` | `org:outlet:manage` | Kelola Outlet | Menambah/mengubah outlet, jam operasional, dan pajak/servis. |
| 11 | `catalog` | `catalog:item:read` | Lihat Menu & Kategori | Melihat katalog produk, varian, modifier, dan kategori. |
| 12 | `catalog` | `catalog:item:manage` | Kelola Menu & Produk | Menambah, mengedit, atau menghapus menu, resep, dan foto produk. |
| 13 | `catalog` | `catalog:price:override` | Penyesuaian Harga Outlet | Menyesuaikan harga produk khusus untuk outlet fisik tertentu. |
| 14 | `catalog` | `catalog:recipe:view_cost` | Lihat HPP & Biaya Resep | Melihat HPP/COGS dan biaya bahan rahasia (*Restricted*). |
| 15 | `pos` | `pos:order:read` | Lihat Transaksi Penjualan | Melihat riwayat transaksi penjualan dan tiket pesanan. |
| 16 | `pos` | `pos:order:create` | Buat Pesanan Baru | Mencatat pesanan baru, tiket dapur, dan print bill sementara. |
| 17 | `pos` | `pos:order:payment` | Terima Pembayaran | Menyelesaikan pembayaran (Tunai, QRIS, EDC, Kartu Debit/Kredit). |
| 18 | `pos` | `pos:order:void` | Void Pesanan / Item | Membatalkan pesanan yang sudah terkirim (*Supervisor Required*). |
| 19 | `pos` | `pos:order:discount_apply` | Berikan Diskon Manual | Menerapkan potongan harga manual atau diskon khusus kasir. |
| 20 | `pos` | `pos:shift:open_close` | Buka / Tutup Shift Kas | Input modal kas awal laci dan rekonsiliasi akhir shift kasir. |
| 21 | `pos` | `pos:drawer:open` | Buka Laci Kasir Manual | Membuka laci uang tanpa transaksi (*Supervisor Required*). |
| 22 | `inventory` | `inventory:stock:read` | Lihat Saldo Stok Bahan | Memeriksa ketersediaan stok bahan baku di gudang/cabang. |
| 23 | `inventory` | `inventory:opname:manage` | Kelola Stok Opname | Melakukan penyesuaian fisik stok harian atau bulanan. |
| 24 | `inventory` | `inventory:transfer:request` | Minta Transfer Stok | Mengajukan mutasi stok antar-cabang atau ke gudang pusat. |
| 25 | `inventory` | `inventory:transfer:approve` | Setujui Mutasi Stok | Otorisasi pengiriman mutasi stok bahan baku. |
| 26 | `inventory` | `inventory:po:create` | Buat Purchase Order (PO) | Mengajukan pesanan pembelian bahan baku ke supplier. |
| 27 | `inventory` | `inventory:po:approve` | Persetujuan PO Pembelian | Otorisasi PO pembelian skala brand atau holding. |
| 28 | `finance` | `finance:report:sales` | Laporan Penjualan | Melihat rekapan omzet harian, mingguan, dan bulanan. |
| 29 | `finance` | `finance:report:pnl` | Laporan Laba Rugi | Mengakses laporan P&L konsolidasi dan HPP operasional holding. |
| 30 | `finance` | `finance:payment_method:manage` | Kelola Metode Pembayaran | Mengatur rekening bank, EDC, dan gateway QRIS. |

---

## 2. Katalog 9 Peran Sistem (System Roles)

Sistem memiliki **9 System Default Roles** yang memiliki `is_system = true`:

| Role Code | Display Name | Tier | UI Lookup Visibility | Cakupan & Karakteristik |
| :--- | :--- | :---: | :---: | :--- |
| **`DEVELOPER`** | **Platform Developer** | `ORGANIZATION` | 🔒 **HIDDEN** | **Full Access & Otorisasi Bypass**. Peran internal vendor platform Hanjo. Tidak pernah ditampilkan di combobox UI Backoffice pelanggan. |
| **`OWNER`** | **Pemilik Bisnis (Owner)** | `ORGANIZATION` | 👁️ Visible | Memiliki hak akses penuh ke seluruh fitur organisasi dan holding. |
| **`FINANCE_HQ`** | **Finance & Akuntansi HQ** | `ORGANIZATION` | 👁️ Visible | Konsolidasi keuangan, rekonsiliasi pembayaran, laporan laba rugi, dan approval PO holding. |
| **`CENTRAL_PURCHASING`** | **Purchasing & Logistik Pusat**| `ORGANIZATION` | 👁️ Visible | Pengadaan bahan baku holding, manajemen gudang pusat, dan transfer stok antar-cabang. |
| **`BRAND_MANAGER`** | **Manajer Brand (Konsep)** | `BRAND` | 👁️ Visible | Bertanggung jawab atas 1 brand: menu, formula resep (BOM), dan promo pemasaran brand. |
| **`BRANCH_MANAGER`** | **Manajer Cabang (Store Mgr)** | `OUTLET` | 👁️ Visible | Bertanggung jawab atas 1 cabang fisik: operasional, shift, opname lokal, dan PIN supervisor. |
| **`CASHIER`** | **Kasir POS** | `OUTLET` | 👁️ Visible | Operasional kasir garis depan: transaksi kasir, pembayaran, dan buka/tutup shift kas. |
| **`BARISTA` / `KITCHEN`** | **Barista / Tim Dapur** | `OUTLET` | 👁️ Visible | Kitchen Display System (KDS): melihat tiket pesanan dan memperbarui status pengerjaan makanan/minuman. |
| **`WAITER`** | **Pelayan Meja (Waiter)** | `OUTLET` | 👁️ Visible | Mencatat pesanan meja tamu (*table order*) dan mengecek status ketersediaan meja. |

---

## 3. Matriks Hak Akses (Role-Permission Matrix)

**Keterangan Simbol:**
* `✅` : Memiliki izin langsung (*Granted*).
* `🔑` : Dapat dieksekusi dengan otorisasi **Supervisor PIN Override** (`X-Supervisor-PIN`).
* `❌` : Tidak memiliki izin (*Denied*).
* `⚡` : **Bypass Otorisasi Penuh** (Role `DEVELOPER`).

| Kode Permission | DEV | OWNER | FIN | PURCH | BRAND_MGR | STORE_MGR | CASHIER | BARISTA | WAITER |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **User & RBAC** | | | | | | | | | |
| `user:staff:read` | ⚡ | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `user:staff:create` | ⚡ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `user:staff:update` | ⚡ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `user:staff:delete` | ⚡ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `rbac:role:read` | ⚡ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `rbac:role:manage` | ⚡ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Organization** | | | | | | | | | |
| `org:brand:read` | ⚡ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| `org:brand:manage` | ⚡ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `org:outlet:read` | ⚡ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| `org:outlet:manage` | ⚡ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Menu & Catalog** | | | | | | | | | |
| `catalog:item:read` | ⚡ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `catalog:item:manage` | ⚡ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| `catalog:price:override`| ⚡ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| `catalog:recipe:view_cost`| ⚡| ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Point of Sale (POS)** | | | | | | | | | |
| `pos:order:read` | ⚡ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `pos:order:create` | ⚡ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ |
| `pos:order:payment` | ⚡ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ |
| `pos:order:void` | ⚡ | ✅ | ❌ | ❌ | ❌ | ✅ | 🔑 | ❌ | ❌ |
| `pos:order:discount_apply`| ⚡| ✅ | ❌ | ❌ | ❌ | ✅ | 🔑 | ❌ | ❌ |
| `pos:shift:open_close` | ⚡ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ |
| `pos:drawer:open` | ⚡ | ✅ | ❌ | ❌ | ❌ | ✅ | 🔑 | ❌ | ❌ |
| **Inventory** | | | | | | | | | |
| `inventory:stock:read` | ⚡ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| `inventory:opname:manage`| ⚡| ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `inventory:transfer:request`| ⚡| ✅ | ❌ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `inventory:transfer:approve`| ⚡| ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `inventory:po:create` | ⚡ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `inventory:po:approve` | ⚡ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Finance & Reports** | | | | | | | | | |
| `finance:report:sales` | ⚡ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| `finance:report:pnl` | ⚡ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `finance:payment_method:manage`| ⚡| ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

---

## 4. Ketentuan Teknis Peran `DEVELOPER`

1. **Bypass Otorisasi**:
   * Pada middleware otorisasi:
     ```go
     if contextutil.GetUserRole(c) == "DEVELOPER" {
         return c.Next() // Bypass all permission & tier guards
     }
     ```
2. **Pengecualian dari Lookup (`GET /api/v1/roles/lookup`)**:
   * Query database lookup combobox wajib memfilter:
     ```sql
     WHERE code != 'DEVELOPER' AND status = 'ACTIVE'
     ```
   * Dengan filter ini, pelanggan/Owner tidak dapat melihat dan tidak dapat meng-assign role `DEVELOPER` kepada siapapun.
3. **Kekebalan Modifikasi (Immutability)**:
   * Role `DEVELOPER` memiliki flag `is_system = true`.
   * Endpoint `PUT /api/v1/roles/:id` dan `DELETE /api/v1/roles/:id` dilarang memodifikasi role dengan code `DEVELOPER`.
