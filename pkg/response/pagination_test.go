package response_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/hanjo-erp/hanjo/pkg/response"
)

func TestNewMeta(t *testing.T) {
	meta := response.NewMeta(1, 20, 95)
	if meta.Pagination == nil {
		t.Fatal("expected non-nil pagination meta")
	}
	if meta.Pagination.Page != 1 || meta.Pagination.PerPage != 20 || meta.Pagination.TotalData != 95 || meta.Pagination.TotalPages != 5 {
		t.Errorf("unexpected meta calculation: %+v", meta.Pagination)
	}

	metaZero := response.NewMeta(1, 0, 0)
	if metaZero.Pagination.TotalPages != 0 {
		t.Errorf("expected 0 total pages for 0 limit, got %d", metaZero.Pagination.TotalPages)
	}

	meta.WithSummary(map[string]any{"total_revenue": 100000})
	if meta.Summary == nil {
		t.Error("expected non-nil summary")
	}
}

func TestPaginationParams_Offset(t *testing.T) {
	p := response.PaginationParams{Page: 3, PerPage: 15}
	if p.Offset() != 30 {
		t.Errorf("expected offset 30, got %d", p.Offset())
	}
}

func TestPaginationParams_PaginateScope(t *testing.T) {
	p := response.PaginationParams{Page: 2, PerPage: 10}
	scope := p.Paginate()
	if scope == nil {
		t.Fatal("expected non-nil scope")
	}

	aliasScope := p.Scope()
	if aliasScope == nil {
		t.Fatal("expected non-nil alias scope")
	}

	standaloneScope := response.Paginate(p)
	if standaloneScope == nil {
		t.Fatal("expected non-nil standalone scope")
	}

	rawScope := response.PaginateScope(2, 10)
	if rawScope == nil {
		t.Fatal("expected non-nil raw scope")
	}
}

func TestPaginationParams_SortScope(t *testing.T) {
	p1 := response.PaginationParams{SortBy: ""}
	scope1 := p1.Sort("created_at DESC")
	if scope1 == nil {
		t.Fatal("expected non-nil scope for default sort")
	}

	p2 := response.PaginationParams{SortBy: "username", SortOrder: "desc"}
	scope2 := p2.Sort("created_at DESC")
	if scope2 == nil {
		t.Fatal("expected non-nil scope for explicit sort")
	}
}

func TestParsePagination(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		p := response.ParsePagination(c)
		return c.JSON(p)
	})

	t.Run("default params", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed test request: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("query params with limit fallback", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?page=3&limit=50&sort_by=email&sort_order=asc", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed test request: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})
}
