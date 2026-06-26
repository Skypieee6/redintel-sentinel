package discovery

import (
	"context"
	"errors"
	"testing"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

type fakeSource struct {
	name     string
	supports map[models.AssetType]bool
	findings []Finding
	err      error
}

func (f fakeSource) Name() string                     { return f.name }
func (f fakeSource) Supports(t models.AssetType) bool { return f.supports[t] }
func (f fakeSource) Discover(context.Context, Input) ([]Finding, error) {
	return f.findings, f.err
}

func TestEngineAggregatesAndDeduplicates(t *testing.T) {
	a := fakeSource{
		name:     "a",
		supports: map[models.AssetType]bool{models.AssetDomain: true},
		findings: []Finding{
			{Type: models.AssetSubdomain, Value: "www.example.com"},
			{Type: models.AssetSubdomain, Value: "WWW.example.com"}, // dup (case-insensitive)
		},
	}
	b := fakeSource{
		name:     "b",
		supports: map[models.AssetType]bool{models.AssetDomain: true},
		findings: []Finding{
			{Type: models.AssetSubdomain, Value: "www.example.com"}, // dup across sources
			{Type: models.AssetCertificate, Value: "CN=example.com"},
		},
	}

	eng := New(100, a, b)
	out, err := eng.Run(context.Background(), Input{Type: models.AssetDomain, Value: "example.com"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 deduped findings, got %d: %+v", len(out), out)
	}
	for _, f := range out {
		if f.Attributes == nil {
			t.Fatalf("attributes should be initialized, got nil for %q", f.Value)
		}
	}
}

func TestEngineSkipsUnsupportedSources(t *testing.T) {
	only := fakeSource{
		name:     "cidr-only",
		supports: map[models.AssetType]bool{models.AssetCIDR: true},
		findings: []Finding{{Type: models.AssetDNSRecord, Value: "10.0.0.1 PTR host"}},
	}
	eng := New(100, only)
	if _, err := eng.Run(context.Background(), Input{Type: models.AssetDomain, Value: "example.com"}); err == nil {
		t.Fatal("expected error when no source supports the input type")
	}
}

func TestEngineToleratesPartialFailure(t *testing.T) {
	failing := fakeSource{name: "bad", supports: map[models.AssetType]bool{models.AssetDomain: true}, err: errors.New("boom")}
	ok := fakeSource{name: "good", supports: map[models.AssetType]bool{models.AssetDomain: true},
		findings: []Finding{{Type: models.AssetSubdomain, Value: "api.example.com"}}}

	out, err := New(100, failing, ok).Run(context.Background(), Input{Type: models.AssetDomain, Value: "example.com"})
	if err != nil {
		t.Fatalf("partial failure should not fail the run: %v", err)
	}
	if len(out) != 1 || out[0].Value != "api.example.com" {
		t.Fatalf("unexpected findings: %+v", out)
	}
}

func TestEngineRejectsEmptyInput(t *testing.T) {
	if _, err := New(100).Run(context.Background(), Input{Type: models.AssetDomain, Value: "  "}); err == nil {
		t.Fatal("expected error for empty input value")
	}
}

func TestValidDiscoveryInput(t *testing.T) {
	for _, ok := range []models.AssetType{models.AssetDomain, models.AssetSubdomain, models.AssetASN, models.AssetCIDR} {
		if !models.ValidDiscoveryInput(ok) {
			t.Errorf("%q should be a valid discovery input", ok)
		}
	}
	for _, bad := range []models.AssetType{models.AssetIP, models.AssetTechnology, models.AssetType("nope")} {
		if models.ValidDiscoveryInput(bad) {
			t.Errorf("%q should not be a valid discovery input", bad)
		}
	}
}

func TestNormalizeHost(t *testing.T) {
	cases := map[string]string{
		"HTTPS://Example.com/": "example.com",
		"*.example.com":        "example.com",
		"sub.example.com.":     "sub.example.com",
		"  api.example.com  ":  "api.example.com",
	}
	for in, want := range cases {
		if got := normalizeHost(in); got != want {
			t.Errorf("normalizeHost(%q) = %q, want %q", in, got, want)
		}
	}
}
