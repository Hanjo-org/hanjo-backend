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
