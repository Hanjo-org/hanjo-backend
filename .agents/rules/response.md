# HTTP Response Guidelines

All Delivery layer HTTP handlers **MUST** use `pkg/response`. Never call `c.JSON(fiber.Map{...})` directly.

## 1. Envelope Structure

```json
{
  "success": true,
  "message": "resource created successfully",
  "data": { ... },
  "meta": {
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total_data": 50,
      "total_pages": 3
    },
    "summary": { ... }
  },
  "error": null
}
```

*   `data`: Always present in the payload.
*   `message`: Short human-readable summary (omitted if empty).
*   `meta`: Present on paginated list and lookup endpoints. Contains nested `pagination` and optional `summary`.
*   `error`: Present on failure (`{"code": "...", "message": "...", "details": { ... }}}`).

## 2. Handler Helpers

*   `response.Success(c, statusCode, message, data)`: For single-item or non-paginated endpoints.
*   `response.SuccessWithMeta(c, statusCode, message, data, meta)`: For paginated list endpoints.
*   `response.Fail(c, statusCode, code, message)`: For mapped errors.
*   `response.FailWithDetails(c, statusCode, code, message, details)`: For validation failures with per-field errors.

## 3. List Endpoint Pattern

```go
func (h *OrderHandler) List(c fiber.Ctx) error {
    params := response.ParsePagination(c)

    orders, total, err := h.usecase.ListOrders(c.Context(), params)
    if err != nil {
        return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list orders")
    }

    meta := response.NewMeta(params.Page, params.PerPage, total)
    return response.SuccessWithMeta(c, fiber.StatusOK, "", orders, meta)
}
```
