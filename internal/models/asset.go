package models

import "time"

// AssetType enumerates the kinds of attack-surface assets tracked.
type AssetType string

const (
	AssetDomain      AssetType = "domain"
	AssetSubdomain   AssetType = "subdomain"
	AssetIP          AssetType = "ip"
	AssetCIDR        AssetType = "cidr"
	AssetASN         AssetType = "asn"
	AssetDNSRecord   AssetType = "dns_record"
	AssetCertificate AssetType = "certificate"
	AssetTechnology  AssetType = "technology"
)

var validAssetTypes = map[AssetType]bool{
	AssetDomain: true, AssetSubdomain: true, AssetIP: true, AssetCIDR: true,
	AssetASN: true, AssetDNSRecord: true, AssetCertificate: true, AssetTechnology: true,
}

// Valid reports whether the asset type is recognized.
func (t AssetType) Valid() bool { return validAssetTypes[t] }

// Asset is a single attack-surface element belonging to a project. Heterogeneous
// per-type fields are stored in Attributes (JSONB), keeping the schema flexible
// across domains, IPs, certificates, technologies and more.
type Asset struct {
	ID         string         `json:"id"`
	OrgID      string         `json:"org_id"`
	ProjectID  string         `json:"project_id"`
	Type       AssetType      `json:"type"`
	Value      string         `json:"value"`
	Tags       []string       `json:"tags"`
	Attributes map[string]any `json:"attributes"`
	Status     string         `json:"status"`
	FirstSeen  time.Time      `json:"first_seen"`
	LastSeen   time.Time      `json:"last_seen"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// AssetTypeCount is a per-type asset tally.
type AssetTypeCount struct {
	Type  AssetType `json:"type"`
	Count int       `json:"count"`
}

// ProjectAssetCount tallies assets within a project.
type ProjectAssetCount struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Assets    int    `json:"assets"`
}

// DashboardSummary aggregates key ASM metrics for an organization.
type DashboardSummary struct {
	TotalAssets   int               `json:"total_assets"`
	AssetsByType  []AssetTypeCount  `json:"assets_by_type"`
	RecentChanges []Asset           `json:"recent_changes"`
	ProjectStats  ProjectStatistics `json:"project_statistics"`
	TeamStats     TeamStatistics    `json:"team_statistics"`
}

// ProjectStatistics summarizes projects in an organization.
type ProjectStatistics struct {
	Total     int                 `json:"total"`
	Active    int                 `json:"active"`
	Archived  int                 `json:"archived"`
	ByProject []ProjectAssetCount `json:"by_project"`
}

// TeamStatistics summarizes teams and membership in an organization.
type TeamStatistics struct {
	Teams          int `json:"teams"`
	Members        int `json:"members"`
	PendingInvites int `json:"pending_invites"`
}
