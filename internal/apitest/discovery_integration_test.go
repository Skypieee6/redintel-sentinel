package apitest

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Skypieee6/redintel-sentinel/internal/discovery"
	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// fakeDiscoveryEngine returns deterministic findings so the discovery workflow
// can be exercised offline (no DNS or Certificate Transparency dependency).
type fakeDiscoveryEngine struct{}

func (fakeDiscoveryEngine) Run(_ context.Context, in discovery.Input) ([]discovery.Finding, error) {
	return []discovery.Finding{
		{Type: models.AssetSubdomain, Value: "www." + in.Value, Source: "fake"},
		{Type: models.AssetSubdomain, Value: "api." + in.Value, Source: "fake"},
		{Type: models.AssetDNSRecord, Value: in.Value + " A 93.184.216.34", Source: "fake",
			Attributes: map[string]any{"record_type": "A", "value": "93.184.216.34"}},
		{Type: models.AssetCertificate, Value: "CN=" + in.Value, Source: "fake",
			Attributes: map[string]any{"issuer": "Test CA"}},
	}, nil
}

func TestPassiveDiscoveryWorkflow(t *testing.T) {
	engine, services := setupServices(t)
	services.Discovery.SetEngine(fakeDiscoveryEngine{})

	c := &apiClient{t: t, engine: engine}
	email := fmt.Sprintf("disc_%d@test.local", time.Now().UnixNano())

	// Register
	code, body := c.do(http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": "supersecret", "full_name": "Discovery User",
	}, nil)
	if code != http.StatusCreated {
		t.Fatalf("register: %d %v", code, body)
	}
	c.token = data(body)["tokens"].(map[string]any)["access_token"].(string)

	// Org + project
	code, body = c.do(http.MethodPost, "/api/v1/orgs", map[string]any{"name": "Discovery Org"}, nil)
	if code != http.StatusCreated {
		t.Fatalf("create org: %d %v", code, body)
	}
	orgID := data(body)["id"].(string)

	code, body = c.do(http.MethodPost, "/api/v1/orgs/"+orgID+"/projects", map[string]any{"name": "Recon"}, nil)
	if code != http.StatusCreated {
		t.Fatalf("create project: %d %v", code, body)
	}
	projectID := data(body)["id"].(string)
	base := "/api/v1/orgs/" + orgID + "/projects/" + projectID

	// Invalid discovery input type rejected.
	if code, _ := c.do(http.MethodPost, base+"/discovery", map[string]any{
		"input_type": "technology", "input_value": "x",
	}, nil); code != http.StatusBadRequest {
		t.Fatalf("invalid discovery input type status = %d, want 400", code)
	}

	// Start discovery (accepted -> async).
	code, body = c.do(http.MethodPost, base+"/discovery", map[string]any{
		"input_type": "domain", "input_value": "example.com",
	}, nil)
	if code != http.StatusAccepted {
		t.Fatalf("start discovery: %d %v", code, body)
	}
	jobID := data(body)["id"].(string)
	if data(body)["status"] != string(models.DiscoveryStatusPending) {
		t.Fatalf("new job status = %v, want pending", data(body)["status"])
	}

	// Poll the job until it reaches a terminal state.
	var job map[string]any
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		code, body = c.do(http.MethodGet, base+"/discovery/"+jobID, nil, nil)
		if code != http.StatusOK {
			t.Fatalf("get discovery job: %d %v", code, body)
		}
		job = data(body)
		if job["status"] == string(models.DiscoveryStatusCompleted) ||
			job["status"] == string(models.DiscoveryStatusFailed) {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}

	if job["status"] != string(models.DiscoveryStatusCompleted) {
		t.Fatalf("discovery job did not complete: %v", job)
	}
	if int(job["assets_found"].(float64)) != 4 {
		t.Fatalf("expected 4 assets found, got %v", job["assets_found"])
	}
	if int(job["assets_created"].(float64)) != 4 {
		t.Fatalf("expected 4 assets created on first run, got %v", job["assets_created"])
	}
	results, ok := job["results"].([]any)
	if !ok || len(results) != 4 {
		t.Fatalf("expected 4 results, got %v", job["results"])
	}

	// Findings must have become normal asset records.
	code, body = c.do(http.MethodGet, base+"/assets?type=subdomain&q=www", nil, nil)
	if code != http.StatusOK {
		t.Fatalf("list assets: %d %v", code, body)
	}
	if int(data(body)["total"].(float64)) < 1 {
		t.Fatalf("discovered subdomain should appear in inventory, got total %v", data(body)["total"])
	}

	// Discovery history lists the job.
	code, body = c.do(http.MethodGet, base+"/discovery", nil, nil)
	if code != http.StatusOK {
		t.Fatalf("list discovery jobs: %d %v", code, body)
	}
	jobs, ok := data(body)["jobs"].([]any)
	if !ok || len(jobs) < 1 {
		t.Fatalf("expected discovery history to contain the job, got %v", data(body)["jobs"])
	}

	// Re-running the same seed should refresh, not duplicate, the assets.
	code, body = c.do(http.MethodPost, base+"/discovery", map[string]any{
		"input_type": "domain", "input_value": "example.com",
	}, nil)
	if code != http.StatusAccepted {
		t.Fatalf("second discovery: %d %v", code, body)
	}
	jobID2 := data(body)["id"].(string)
	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		_, body = c.do(http.MethodGet, base+"/discovery/"+jobID2, nil, nil)
		job = data(body)
		if job["status"] == string(models.DiscoveryStatusCompleted) {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if job["status"] != string(models.DiscoveryStatusCompleted) {
		t.Fatalf("second discovery did not complete: %v", job)
	}
	if int(job["assets_created"].(float64)) != 0 {
		t.Fatalf("re-running discovery should create 0 new assets, got %v", job["assets_created"])
	}
}
