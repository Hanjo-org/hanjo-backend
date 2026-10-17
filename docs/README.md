# Hanjo ERP & POS - API Documentation (OpenAPI 3.0.3 / Swagger)

Dokumentasi RESTful API untuk **Hanjo ERP & POS Backend** dirancang modular per domain bisnis untuk mempermudah integrasi tim frontend (Dashboard Web & Tablet POS):

```text
docs/
├── docs.go                  # Go embed.FS (seluruh berkas dokumentasi di-embed ke binary aplikasi)
├── index.html               # Interactive Swagger UI kustom dengan multi-spec dropdown switcher
├── swagger.yaml             # Master API Specification (seluruh endpoint sistem)
├── modules/
│   ├── auth.yaml            # Modul Otentikasi, POS PIN, JWT RS256, & Session
│   └── organization.yaml    # Arsitektur & Model Hierarki Multi-Tenancy 3-Tier
└── README.md                # Panduan dokumentasi ini
```

---

## 🌐 Akses Dokumentasi Interaktif (Browser)

Ketika server backend berjalan (`make dev` atau `make run` pada port default `8080`):

* **Interactive Swagger UI**: [http://localhost:8080/docs](http://localhost:8080/docs) *(atau alias [/swagger](http://localhost:8080/swagger))*
  * Dilengkapi dropdown selector di topbar untuk berganti antar spesifikasi:
    1. **🌟 Master Spec (All Features)**
    2. **🔐 Auth & Session Module**
    3. **🏢 Multi-Tenancy Hierarchy (3-Tier)**
  * Mendukung tombol **"Try it out"** langsung untuk menguji request HTTP.
  * Mendukung tombol **"Authorize"** dengan memasukkan Access Token JWT (`Bearer <access_token>`).
* **Master Specification (YAML)**: [http://localhost:8080/docs/swagger.yaml](http://localhost:8080/docs/swagger.yaml)
* **Auth Module Spec (YAML)**: [http://localhost:8080/docs/modules/auth.yaml](http://localhost:8080/docs/modules/auth.yaml)
* **Organization Spec (YAML)**: [http://localhost:8080/docs/modules/organization.yaml](http://localhost:8080/docs/modules/organization.yaml)

---

## 🚀 Cara Import ke Tooling API

### 1. Postman / Insomnia
- **Import Seluruh Endpoint**: Import file `docs/swagger.yaml` atau paste URL `http://localhost:8080/docs/swagger.yaml`.
- **Import Spesifik Modul**: Tarik langsung file `docs/modules/auth.yaml` jika tim hanya berfokus pada integrasi autentikasi.

### 2. Swagger Editor (Online)
Buka [Swagger Editor](https://editor.swagger.io) $\rightarrow$ Pilih **File** $\rightarrow$ **Import File** $\rightarrow$ Unggah file YAML yang diinginkan.

---

## 🔑 Akun Uji Coba Demo (Database Seeder)

Untuk mencoba langsung via tombol **"Try it out"** di Swagger UI, gunakan kredensial demo hasil `make seed`:

| Peran / Akun | Identifier | Password / PIN | Scope Akses |
| :--- | :--- | :--- | :--- |
| **Developer / Super Admin** | `dev@hanjo.id` / `developer` | `Password123!` | Global (Full Access) |
| **Owner (Holding HQ)** | `owner@harmoni.id` / `owner` | `Password123!` | Organization HQ |
| **Brand Manager** | `manager.daily@harmoni.id` | `Password123!` | Brand (`Daily Coffee`) |
| **Branch Manager** | `manager.kemang@harmoni.id` | `Password123!` | Outlet (`Daily Coffee - Kemang`) |
| **Kasir POS Kemang** | Outlet ID: `04826e26-f9e9-4200-b653-18b2b652c13c`<br>Username: `kasir.kemang` | PIN: `123456` | POS Tablet Kemang |
| **Barista Kemang** | Outlet ID: `04826e26-f9e9-4200-b653-18b2b652c13c`<br>Username: `barista.kemang` | PIN: `654321` | Kitchen / Bar Display Kemang |
