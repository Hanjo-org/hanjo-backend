# Data Modeling & Entity Guidelines

Conventions for designing Domain Entities and GORM Persistence Models conforming to Strict Clean Architecture.

## 1. Domain Entities (`internal/modules/[module]/domain/`)

*   **Zero Third-Party DB Dependencies**: The `domain/` layer must NEVER import `gorm.io/gorm` or use `gorm:"..."` struct tags.
*   **Primary Keys**: Use **UUID v4** (string representation) generated inside entity constructors using `github.com/google/uuid`.
*   **Timestamps & Soft Delete**: Standard audit fields use native standard library types:
    ```go
    type Product struct {
        ID        string     `json:"id"`
        Name      string     `json:"name"`
        Price     int64      `json:"price"`
        CreatedAt time.Time  `json:"created_at"`
        UpdatedAt time.Time  `json:"updated_at"`
        DeletedAt *time.Time `json:"-"`
    }
    ```
*   **Constructors**: Use constructor functions initializing default UUIDs and timestamps:
    ```go
    func NewProduct(name string, price int64) *Product {
        now := time.Now()
        return &Product{
            ID:        uuid.NewString(),
            Name:      name,
            Price:     price,
            CreatedAt: now,
            UpdatedAt: now,
        }
    }
    ```

## 2. GORM Persistence Models (`internal/modules/[module]/infrastructure/`)

*   Define database models in `infrastructure/<entity>_model.go` named `<Entity>Model`.
*   **GORM Tags**: Use explicit lowercase `column:name` tags and SQL types:
    ```go
    type ProductModel struct {
        ID        string         `gorm:"column:id;type:varchar(36);primaryKey"`
        Name      string         `gorm:"column:name;type:varchar(255);not null"`
        Price     int64          `gorm:"column:price;not null"`
        CreatedAt time.Time      `gorm:"column:created_at;not null"`
        UpdatedAt time.Time      `gorm:"column:updated_at;not null"`
        DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
    }

    func (ProductModel) TableName() string {
        return "products"
    }
    ```
*   **Mappers**: Every model provides bidirectional mapping:
    - `(m *ProductModel) ToDomain() *domain.Product`
    - `FromDomain(p *domain.Product) *ProductModel`
*   **Database Migrations**: Schema changes must be versioned in `migrations/` via `golang-migrate` (see `.agents/skills/db-migration/SKILL.md`). Do NOT use GORM AutoMigrate.
