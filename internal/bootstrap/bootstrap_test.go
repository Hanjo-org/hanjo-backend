package bootstrap_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/hanjo-erp/hanjo/internal/bootstrap"
	"github.com/hanjo-erp/hanjo/internal/config"
)

func TestEndpoints(t *testing.T) {
	cfg := &config.Config{
		AppName: "hanjo-backend",
		AppEnv:  "test",
	}
	logger := zap.NewNop()

	app := bootstrap.SetupApp(cfg, nil, logger)

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		{
			name:         "Health Check",
			url:          "/health",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Root Endpoint",
			url:          "/",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Swagger UI HTML",
			url:          "/docs",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Master Swagger YAML",
			url:          "/docs/swagger.yaml",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Auth Module YAML",
			url:          "/docs/modules/auth.yaml",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Organization Module YAML",
			url:          "/docs/modules/organization.yaml",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Swagger Alias Redirect",
			url:          "/swagger",
			expectedCode: http.StatusFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCode, resp.StatusCode)
		})
	}
}
