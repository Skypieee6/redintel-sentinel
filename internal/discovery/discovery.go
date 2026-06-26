// Package discovery implements passive, defensive attack-surface discovery.
//
// It is strictly limited to passive reconnaissance against authorized scopes:
// public DNS resolution and public Certificate Transparency logs. It performs
// no intrusive scanning, brute forcing or exploitation.
package discovery

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Skypieee6/redintel-sentinel/internal/models"
)

// Input is a discovery seed: an authorized asset to enumerate from.
type Input struct {
	Type  models.AssetType
	Value string
}

// Finding is a single passively discovered asset, normalized into the same
// shape as inventory assets.
type Finding struct {
	Type       models.AssetType
	Value      string
	Source     string
	Attributes map[string]any
}

// Source is a single passive discovery technique (e.g. DNS, Certificate
// Transparency). Sources are pluggable so the engine can be composed and the
// service can inject deterministic fakes in tests.
type Source interface {
	Name() string
	Supports(t models.AssetType) bool
	Discover(ctx context.Context, in Input) ([]Finding, error)
}

// Engine runs a set of sources over an input and returns deduplicated findings.
type Engine interface {
	Run(ctx context.Context, in Input) ([]Finding, error)
}

type engine struct {
	sources     []Source
	maxFindings int
}

// New builds an engine from the given sources, capping the number of findings.
func New(maxFindings int, sources ...Source) Engine {
	if maxFindings <= 0 {
		maxFindings = 2000
	}
	return &engine{sources: sources, maxFindings: maxFindings}
}

// Run executes every source that supports the input type, aggregating and
// deduplicating findings. Individual source errors are tolerated as long as at
// least one source produces results; if all supporting sources fail, the joined
// error is returned.
func (e *engine) Run(ctx context.Context, in Input) ([]Finding, error) {
	if strings.TrimSpace(in.Value) == "" {
		return nil, errors.New("discovery input value is empty")
	}

	seen := make(map[string]bool)
	var out []Finding
	var errs []error
	supported := false

	for _, src := range e.sources {
		if !src.Supports(in.Type) {
			continue
		}
		supported = true
		findings, err := src.Discover(ctx, in)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", src.Name(), err))
			continue
		}
		for _, f := range findings {
			value := strings.TrimSpace(f.Value)
			if value == "" {
				continue
			}
			key := string(f.Type) + "|" + strings.ToLower(value)
			if seen[key] {
				continue
			}
			seen[key] = true
			f.Value = value
			if f.Attributes == nil {
				f.Attributes = map[string]any{}
			}
			out = append(out, f)
			if len(out) >= e.maxFindings {
				return out, nil
			}
		}
	}

	if !supported {
		return nil, fmt.Errorf("no passive discovery source supports input type %q", in.Type)
	}
	if len(out) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return out, nil
}
