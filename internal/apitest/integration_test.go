package apitest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Skypieee6/redintel-sentinel/internal/auth"
	"github.com/Skypieee6/redintel-sentinel/internal/cache"
	"github.com/Skypieee6/redintel-sentinel/internal/config"
	"github.com/Skypieee6/redintel-sentinel/internal/database"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
	"github.com/Skypieee6/redintel-sentinel/internal/router"
	"github.com/Skypieee6/redintel-sentinel/internal/service"
)

// setup builds a router backed by real Postgres + Redis. It skips the test if
// the datastores are not reachable (e.g. unit-only CI without services).
func setup(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	ctx := context.Background()

	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		t.Skipf("postgres unavailable, skipping integration test: %v", err)
	}
	t.Cleanup(db.Close)

	redis, err := cache.New(ctx, cfg.Redis)
	if err != nil {
		t.Skipf("redis unavailable, skipping integration test: %v", err)
	}
	t.Cleanup(func() { _ = redis.Close() })

	repos := repository.New(db.Pool)
	jwt := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL)
	cfg.Auth.BcryptCost = 4 // speed up tests
	services := service.New(repos, jwt, redis, cfg.Auth, zapNop())

	return router.New(router.Dependencies{
		Config: cfg, Logger: zapNop(), DB: db, Redis: redis,
		Repos: repos, Services: services, JWT: jwt,
	})
}

type apiClient struct {
	t      *testing.T
	engine *gin.Engine
	token  string
}

func (c *apiClient) do(method, path string, body any, headers map[string]string) (int, map[string]any) {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	c.engine.ServeHTTP(rec, req)

	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func data(m map[string]any) map[string]any {
	if d, ok := m["data"].(map[string]any); ok {
		return d
	}
	return m
}

func TestCorePlatformFlow(t *testing.T) {
	engine := setup(t)
	c := &apiClient{t: t, engine: engine}

	email := fmt.Sprintf("user_%d@test.local", time.Now().UnixNano())

	// Register
	code, body := c.do(http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": "supersecret", "full_name": "Test User",
	}, nil)
	if code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %v", code, body)
	}
	d := data(body)
	tokens := d["tokens"].(map[string]any)
	c.token = tokens["access_token"].(string)
	refresh := tokens["refresh_token"].(string)

	// Me
	code, body = c.do(http.MethodGet, "/api/v1/auth/me", nil, nil)
	if code != http.StatusOK || data(body)["email"] != email {
		t.Fatalf("me status = %d, body = %v", code, body)
	}

	// Wrong password
	if code, _ := c.do(http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": email, "password": "nope-nope",
	}, nil); code != http.StatusUnauthorized {
		t.Fatalf("login with wrong password status = %d, want 401", code)
	}

	// Create org
	code, body = c.do(http.MethodPost, "/api/v1/orgs", map[string]any{"name": "Acme Security"}, nil)
	if code != http.StatusCreated {
		t.Fatalf("create org status = %d, body = %v", code, body)
	}
	orgID := data(body)["id"].(string)

	// Create project
	code, body = c.do(http.MethodPost, "/api/v1/orgs/"+orgID+"/projects", map[string]any{
		"name": "External Perimeter", "description": "authorized scope",
	}, nil)
	if code != http.StatusCreated {
		t.Fatalf("create project status = %d, body = %v", code, body)
	}

	// List projects
	code, body = c.do(http.MethodGet, "/api/v1/orgs/"+orgID+"/projects", nil, nil)
	if code != http.StatusOK {
		t.Fatalf("list projects status = %d", code)
	}
	if arr, ok := body["data"].([]any); !ok || len(arr) == 0 {
		t.Fatalf("expected at least one project, got %v", body["data"])
	}

	// Create API key and use it
	code, body = c.do(http.MethodPost, "/api/v1/auth/api-keys", map[string]any{"name": "ci-key"}, nil)
	if code != http.StatusCreated {
		t.Fatalf("create api key status = %d, body = %v", code, body)
	}
	secret := data(body)["secret"].(string)
	if secret == "" {
		t.Fatal("api key secret should be returned once")
	}
	prevToken := c.token
	c.token = ""
	code, body = c.do(http.MethodGet, "/api/v1/auth/me", nil, map[string]string{"X-API-Key": secret})
	if code != http.StatusOK || data(body)["email"] != email {
		t.Fatalf("api-key auth status = %d, body = %v", code, body)
	}
	c.token = prevToken

	// Refresh token rotation
	code, body = c.do(http.MethodPost, "/api/v1/auth/refresh", map[string]any{"refresh_token": refresh}, nil)
	if code != http.StatusOK {
		t.Fatalf("refresh status = %d, body = %v", code, body)
	}
	// Old refresh token must now be invalid (rotation).
	if code, _ := c.do(http.MethodPost, "/api/v1/auth/refresh", map[string]any{"refresh_token": refresh}, nil); code == http.StatusOK {
		t.Fatal("rotated refresh token should no longer be valid")
	}
}

func TestUnauthenticatedRejected(t *testing.T) {
	engine := setup(t)
	c := &apiClient{t: t, engine: engine}
	if code, _ := c.do(http.MethodGet, "/api/v1/orgs", nil, nil); code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated /orgs status = %d, want 401", code)
	}
}
