package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// ReportFormat enumerates supported export formats.
type ReportFormat string

const (
	ReportJSON     ReportFormat = "json"
	ReportCSV      ReportFormat = "csv"
	ReportMarkdown ReportFormat = "markdown"
	ReportHTML     ReportFormat = "html"
)

// Report is a generated export ready to be written to an HTTP response.
type Report struct {
	ContentType string
	Filename    string
	Body        []byte
}

// ReportService generates project asset inventory reports.
type ReportService struct {
	repos *repository.Repositories
}

// Generate produces a report of a project's assets in the requested format.
func (s *ReportService) Generate(ctx context.Context, orgID, projectID string, format ReportFormat) (*Report, error) {
	project, err := s.repos.Projects.Get(ctx, orgID, projectID)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}
	assets, _, err := s.repos.Assets.List(ctx, repository.AssetFilter{
		ProjectID: projectID, Limit: 200, Offset: 0,
	})
	if err != nil {
		return nil, err
	}
	// Page through all assets so the report is complete.
	for len(assets)%200 == 0 && len(assets) > 0 {
		next, _, err := s.repos.Assets.List(ctx, repository.AssetFilter{
			ProjectID: projectID, Limit: 200, Offset: len(assets),
		})
		if err != nil {
			return nil, err
		}
		if len(next) == 0 {
			break
		}
		assets = append(assets, next...)
	}

	switch format {
	case ReportJSON:
		return s.json(project, assets)
	case ReportCSV:
		return s.csv(project, assets)
	case ReportMarkdown:
		return s.markdown(project, assets)
	case ReportHTML:
		return s.html(project, assets)
	default:
		return nil, wrap(ErrValidation, "unsupported report format %q (use json, csv, markdown or html)", format)
	}
}

func (s *ReportService) json(p *models.Project, assets []models.Asset) (*Report, error) {
	payload := map[string]any{
		"report":       "asset_inventory",
		"generated_at": time.Now().UTC(),
		"project":      p,
		"asset_count":  len(assets),
		"assets":       assets,
	}
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}
	return &Report{ContentType: "application/json", Filename: filename(p, "json"), Body: body}, nil
}

func (s *ReportService) csv(p *models.Project, assets []models.Asset) (*Report, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "type", "value", "status", "tags", "first_seen", "last_seen"})
	for _, a := range assets {
		_ = w.Write([]string{
			a.ID, string(a.Type), a.Value, a.Status,
			strings.Join(a.Tags, "|"),
			a.FirstSeen.UTC().Format(time.RFC3339),
			a.LastSeen.UTC().Format(time.RFC3339),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return &Report{ContentType: "text/csv", Filename: filename(p, "csv"), Body: buf.Bytes()}, nil
}

func (s *ReportService) markdown(p *models.Project, assets []models.Asset) (*Report, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# Asset Inventory — %s\n\n", p.Name)
	fmt.Fprintf(&b, "- Generated: %s\n- Assets: %d\n\n", time.Now().UTC().Format(time.RFC3339), len(assets))
	b.WriteString("| Type | Value | Status | Tags | Last Seen |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	for _, a := range assets {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n",
			a.Type, a.Value, a.Status, strings.Join(a.Tags, ", "),
			a.LastSeen.UTC().Format("2006-01-02"))
	}
	return &Report{ContentType: "text/markdown", Filename: filename(p, "md"), Body: []byte(b.String())}, nil
}

func (s *ReportService) html(p *models.Project, assets []models.Asset) (*Report, error) {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\">")
	fmt.Fprintf(&b, "<title>Asset Inventory — %s</title>", html.EscapeString(p.Name))
	b.WriteString("<style>body{font-family:system-ui,sans-serif;margin:2rem;}table{border-collapse:collapse;width:100%;}th,td{border:1px solid #ddd;padding:.5rem;text-align:left;}th{background:#111;color:#fff;}</style></head><body>")
	fmt.Fprintf(&b, "<h1>Asset Inventory — %s</h1>", html.EscapeString(p.Name))
	fmt.Fprintf(&b, "<p>Generated: %s · Assets: %d</p>", time.Now().UTC().Format(time.RFC3339), len(assets))
	b.WriteString("<table><thead><tr><th>Type</th><th>Value</th><th>Status</th><th>Tags</th><th>Last Seen</th></tr></thead><tbody>")
	for _, a := range assets {
		fmt.Fprintf(&b, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			html.EscapeString(string(a.Type)), html.EscapeString(a.Value),
			html.EscapeString(a.Status), html.EscapeString(strings.Join(a.Tags, ", ")),
			a.LastSeen.UTC().Format("2006-01-02"))
	}
	b.WriteString("</tbody></table></body></html>")
	return &Report{ContentType: "text/html; charset=utf-8", Filename: filename(p, "html"), Body: []byte(b.String())}, nil
}

func filename(p *models.Project, ext string) string {
	return fmt.Sprintf("asset-inventory-%s.%s", p.ID, ext)
}
