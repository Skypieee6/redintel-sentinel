package models

import "time"

// DiscoveryStatus represents the lifecycle state of a passive discovery job.
type DiscoveryStatus string

const (
	DiscoveryStatusPending   DiscoveryStatus = "pending"
	DiscoveryStatusRunning   DiscoveryStatus = "running"
	DiscoveryStatusCompleted DiscoveryStatus = "completed"
	DiscoveryStatusFailed    DiscoveryStatus = "failed"
)

// validDiscoveryInputs enumerates the asset types accepted as a discovery seed.
// Discovery is strictly passive and limited to authorized, defensive use.
var validDiscoveryInputs = map[AssetType]bool{
	AssetDomain:    true,
	AssetSubdomain: true,
	AssetASN:       true,
	AssetCIDR:      true,
}

// ValidDiscoveryInput reports whether t is an accepted discovery seed type.
func ValidDiscoveryInput(t AssetType) bool { return validDiscoveryInputs[t] }

// DiscoveryJob is a single passive discovery run scoped to a project.
type DiscoveryJob struct {
	ID            string            `json:"id"`
	OrgID         string            `json:"org_id"`
	ProjectID     string            `json:"project_id"`
	InputType     AssetType         `json:"input_type"`
	InputValue    string            `json:"input_value"`
	Sources       []string          `json:"sources"`
	Status        DiscoveryStatus   `json:"status"`
	Error         string            `json:"error,omitempty"`
	AssetsFound   int               `json:"assets_found"`
	AssetsCreated int               `json:"assets_created"`
	CreatedBy     string            `json:"created_by"`
	StartedAt     *time.Time        `json:"started_at,omitempty"`
	CompletedAt   *time.Time        `json:"completed_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Results       []DiscoveryResult `json:"results,omitempty"`
}

// DiscoveryResult is a single finding produced by a discovery job. Each result
// is linked to the normal asset record it created or refreshed.
type DiscoveryResult struct {
	ID         string         `json:"id"`
	JobID      string         `json:"job_id"`
	AssetID    string         `json:"asset_id"`
	Type       AssetType      `json:"type"`
	Value      string         `json:"value"`
	Source     string         `json:"source"`
	Attributes map[string]any `json:"attributes"`
	IsNew      bool           `json:"is_new"`
	CreatedAt  time.Time      `json:"created_at"`
}
