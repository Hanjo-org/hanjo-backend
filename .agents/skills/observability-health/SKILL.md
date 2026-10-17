---
name: observability-health
description: >-
  Implement and audit production observability, health check probes, and metrics for Go services.
  Use when adding Kubernetes liveness/readiness probes, database health checks, or Prometheus metrics.
---

# Observability & Health Probes Playbook

This skill establishes standards for application health checks (Kubernetes / Docker probes), dependency verification, and metrics instrumentation.

---

## 1. Health Probe Architecture

For cloud-native deployments (Kubernetes, AWS ECS, Docker Compose), decouple process liveness from dependency readiness:

| Endpoint | Probe Type | Purpose | Failure Consequence |
|---|---|---|---|
| `/health/live` | **Liveness Probe** | Verifies the web server process is responsive and not deadlocked. | Orchestrator restarts the container. |
| `/health/ready` | **Readiness Probe** | Verifies database (PostgreSQL) and external dependencies are connected. | Orchestrator stops routing traffic to this pod until ready. |

---

## 2. Health Check Implementation Recipe

### 2.1 Health Check Delivery Handler (`internal/modules/health/delivery/health_handler.go`)

```go
package delivery

import (
    "context"
    "time"

    "github.com/hanjo-erp/hanjo/pkg/response"
    "github.com/gofiber/fiber/v3"
    "gorm.io/gorm"
)

type HealthHandler struct {
    db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
    return &HealthHandler{db: db}
}

// Live godoc - Liveness probe
func (h *HealthHandler) Live(c fiber.Ctx) error {
    return response.Success(c, fiber.StatusOK, fiber.Map{
        "status": "UP",
    })
}

// Ready godoc - Readiness probe checking dependencies
func (h *HealthHandler) Ready(c fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
    defer cancel()

    checks := fiber.Map{
        "database": "UP",
    }
    isHealthy := true

    // Check Database connection
    sqlDB, err := h.db.DB()
    if err != nil || sqlDB.PingContext(ctx) != nil {
        checks["database"] = "DOWN"
        isHealthy = false
    }

    if !isHealthy {
        return response.Fail(c, fiber.StatusServiceUnavailable, "DEPENDENCY_FAILURE", "One or more dependencies are unavailable", checks)
    }

    return response.Success(c, fiber.StatusOK, fiber.Map{
        "status": "UP",
        "checks": checks,
    })
}
```

### 2.2 Route Registration in Bootstrap
```go
healthHandler := healthdelivery.NewHealthHandler(db)
app.Get("/health/live", healthHandler.Live)
app.Get("/health/ready", healthHandler.Ready)
```

---

## 3. Prometheus Metrics Instrumentation

Expose standard Prometheus metrics for Grafana dashboards:

### 3.1 Metric Types to Track
1. **Request Counter**: Total HTTP requests partitioned by status, method, and route:
   `http_requests_total{method="POST", path="/api/v1/users", status="201"}`
2. **Request Latency**: Duration histogram to compute p50, p95, and p99 latencies:
   `http_request_duration_seconds_bucket{method="GET", path="/api/v1/users", le="0.1"}`
3. **Database Connection Pool Metrics**: Open, idle, and in-use connections from `sqlDB.Stats()`.

### 3.2 Metrics Endpoint Route
Mount metrics at `/metrics` (accessible only internally or protected via network policies in production):
```go
import "github.com/ansrivas/fiberprometheus/v2"

// In Bootstrap SetupApp:
prometheus := fiberprometheus.New("backend-api")
prometheus.RegisterAt(app, "/metrics")
app.Use(prometheus.Middleware)
```

---

## 4. Observability Checklist

- [ ] Is `/health/live` fast, lightweight, and independent of external databases?
- [ ] Does `/health/ready` check PostgreSQL ping with a 2-second context timeout?
- [ ] Does a failed readiness check return `503 Service Unavailable` with details in the error payload?
- [ ] Are sensitive connection strings or credentials excluded from health output?
- [ ] Is `/metrics` restricted from public external internet access in production?
