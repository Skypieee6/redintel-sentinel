package service

import (
	"strings"
	"testing"
	"time"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

func sampleProjectAndAssets() (*models.Project, []models.Asset) {
	p := &models.Project{ID: "p-1", Name: "Acme Perimeter", OrgID: "o-1"}
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	assets := []models.Asset{
		{ID: "a-1", Type: models.AssetDomain, Value: "acme.com", Status: "active", Tags: []string{"prod"}, FirstSeen: now, LastSeen: now},
		{ID: "a-2", Type: models.AssetIP, Value: "203.0.113.4", Status: "active", Tags: nil, FirstSeen: now, LastSeen: now},
	}
	return p, assets
}

func TestReportJSON(t *testing.T) {
	s := &ReportService{}
	p, assets := sampleProjectAndAssets()
	r, err := s.json(p, assets)
	if err != nil {
		t.Fatal(err)
	}
	if r.ContentType != "application/json" {
		t.Errorf("content type = %q", r.ContentType)
	}
	body := string(r.Body)
	if !strings.Contains(body, "acme.com") || !strings.Contains(body, "\"asset_count\": 2") {
		t.Errorf("json body missing expected content: %s", body)
	}
}

func TestReportCSV(t *testing.T) {
	s := &ReportService{}
	p, assets := sampleProjectAndAssets()
	r, err := s.csv(p, assets)
	if err != nil {
		t.Fatal(err)
	}
	body := string(r.Body)
	if !strings.HasPrefix(body, "id,type,value,status,tags,first_seen,last_seen") {
		t.Errorf("csv header missing: %s", body)
	}
	if !strings.Contains(body, "203.0.113.4") {
		t.Errorf("csv missing asset: %s", body)
	}
}

func TestReportMarkdownAndHTML(t *testing.T) {
	s := &ReportService{}
	p, assets := sampleProjectAndAssets()

	md, err := s.markdown(p, assets)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(md.Body), "# Asset Inventory — Acme Perimeter") {
		t.Errorf("markdown missing heading: %s", md.Body)
	}

	h, err := s.html(p, assets)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(h.Body), "<table>") || !strings.Contains(string(h.Body), "acme.com") {
		t.Errorf("html missing table/content")
	}
}
