package service

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/internal/discovery"
	"github.com/Skypieee6/redintel-sentinel/internal/models"
	"github.com/Skypieee6/redintel-sentinel/internal/repository"
)

// DiscoveryJobPage is a paginated discovery job listing.
type DiscoveryJobPage struct {
	Jobs   []models.DiscoveryJob `json:"jobs"`
	Total  int                   `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

// DiscoveryService orchestrates passive, defensive asset discovery jobs.
type DiscoveryService struct {
	repos   *repository.Repositories
	audit   *AuditService
	engine  discovery.Engine
	log     *zap.Logger
	timeout time.Duration
}

// SetEngine overrides the discovery engine. Used by tests to inject a
// deterministic, offline engine.
func (s *DiscoveryService) SetEngine(e discovery.Engine) { s.engine = e }

// canManageProject mirrors the asset write-access rules: org analyst+, project
// owner, or an explicit project member with analyst+.
func (s *DiscoveryService) canManageProject(ctx context.Context, orgID, projectID, userID string, orgRole models.Role) (bool, error) {
	p, err := s.repos.Projects.Get(ctx, orgID, projectID)
	if err != nil {
		return false, mapRepoErr(err, "project")
	}
	if orgRole.AtLeast(models.RoleAnalyst) {
		return true, nil
	}
	if p.OwnerID == userID {
		return true, nil
	}
	members, err := s.repos.Projects.ListMembers(ctx, projectID)
	if err != nil {
		return false, err
	}
	for _, m := range members {
		if m.UserID == userID && m.Role.AtLeast(models.RoleAnalyst) {
			return true, nil
		}
	}
	return false, nil
}

// Start validates input, enforces access, persists a pending job and launches
// its asynchronous execution.
func (s *DiscoveryService) Start(ctx context.Context, actorID, orgID, projectID string, orgRole models.Role, inputType models.AssetType, inputValue, ip string) (*models.DiscoveryJob, error) {
	ok, err := s.canManageProject(ctx, orgID, projectID, actorID, orgRole)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, wrap(ErrForbidden, "analyst role or above is required to run discovery")
	}
	if !models.ValidDiscoveryInput(inputType) {
		return nil, wrap(ErrValidation, "invalid discovery input type %q (use domain, subdomain, asn or cidr)", inputType)
	}
	if strings.TrimSpace(inputValue) == "" {
		return nil, wrap(ErrValidation, "discovery input value is required")
	}

	job, err := s.repos.Discovery.CreateJob(ctx, &models.DiscoveryJob{
		OrgID:      orgID,
		ProjectID:  projectID,
		InputType:  inputType,
		InputValue: strings.TrimSpace(inputValue),
		CreatedBy:  actorID,
	})
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "discovery.started", actorID, orgID, "discovery_job", job.ID, ip,
		map[string]any{"input_type": string(inputType), "input_value": job.InputValue})

	go s.execute(job, actorID, ip)
	return job, nil
}

// execute runs the discovery engine, persists findings as normal assets and
// records the job outcome. It runs on its own bounded context so it survives
// the originating HTTP request.
func (s *DiscoveryService) execute(job *models.DiscoveryJob, actorID, ip string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if err := s.repos.Discovery.MarkRunning(ctx, job.ID); err != nil {
		s.log.Warn("discovery: mark running failed", zap.String("job", job.ID), zap.Error(err))
	}

	findings, err := s.engine.Run(ctx, discovery.Input{Type: job.InputType, Value: job.InputValue})
	if err != nil && len(findings) == 0 {
		s.log.Warn("discovery run failed", zap.String("job", job.ID), zap.Error(err))
		s.audit.Record(ctx, "discovery.failed", actorID, job.OrgID, "discovery_job", job.ID, ip,
			map[string]any{"error": err.Error()})
		_ = s.repos.Discovery.MarkFailed(ctx, job.ID, err.Error())
		return
	}

	found, created := 0, 0
	results := make([]*models.DiscoveryResult, 0, len(findings))
	for _, f := range findings {
		asset, isNew, aerr := s.repos.Assets.Upsert(ctx, &models.Asset{
			OrgID:      job.OrgID,
			ProjectID:  job.ProjectID,
			Type:       f.Type,
			Value:      f.Value,
			Tags:       []string{"discovered"},
			Attributes: f.Attributes,
		})
		if aerr != nil {
			s.log.Warn("discovery: upsert asset failed", zap.String("value", f.Value), zap.Error(aerr))
			continue
		}
		found++
		if isNew {
			created++
		}
		results = append(results, &models.DiscoveryResult{
			JobID:      job.ID,
			AssetID:    asset.ID,
			Type:       f.Type,
			Value:      f.Value,
			Source:     f.Source,
			Attributes: f.Attributes,
			IsNew:      isNew,
		})
	}

	if err := s.repos.Discovery.AddResults(ctx, results); err != nil {
		s.log.Warn("discovery: persist results failed", zap.String("job", job.ID), zap.Error(err))
	}
	s.audit.Record(ctx, "discovery.completed", actorID, job.OrgID, "discovery_job", job.ID, ip,
		map[string]any{"assets_found": found, "assets_created": created})
	// MarkCompleted is the final write so a terminal status guarantees the job
	// goroutine has finished touching the database.
	if err := s.repos.Discovery.MarkCompleted(ctx, job.ID, found, created); err != nil {
		s.log.Warn("discovery: mark completed failed", zap.String("job", job.ID), zap.Error(err))
	}
}

// ListJobs returns a paginated discovery history for a project.
func (s *DiscoveryService) ListJobs(ctx context.Context, orgID, projectID string, limit, offset int) (*DiscoveryJobPage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	jobs, total, err := s.repos.Discovery.ListJobs(ctx, orgID, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	if jobs == nil {
		jobs = []models.DiscoveryJob{}
	}
	return &DiscoveryJobPage{Jobs: jobs, Total: total, Limit: limit, Offset: offset}, nil
}

// Get returns a discovery job together with its findings.
func (s *DiscoveryService) Get(ctx context.Context, orgID, projectID, id string) (*models.DiscoveryJob, error) {
	job, err := s.repos.Discovery.GetJob(ctx, orgID, projectID, id)
	if err != nil {
		return nil, mapRepoErr(err, "discovery job")
	}
	results, err := s.repos.Discovery.ListResults(ctx, job.ID)
	if err != nil {
		return nil, err
	}
	if results == nil {
		results = []models.DiscoveryResult{}
	}
	job.Results = results
	return job, nil
}
