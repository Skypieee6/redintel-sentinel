// Package version exposes build-time metadata for the application.
//
// The values are injected at build time via -ldflags, e.g.:
//
//	go build -ldflags "-X github.com/Skypieee6/redintel-sentinel/internal/version.Version=1.2.3"
package version

import (
	"fmt"
	"runtime"
)

// Build metadata. Overridden at build time via -ldflags; sensible defaults are
// used for local development builds.
var (
	// Version is the semantic version of the build.
	Version = "dev"
	// Commit is the git commit SHA the build was produced from.
	Commit = "none"
	// BuildTime is the UTC timestamp the build was produced at.
	BuildTime = "unknown"
)

// Info is a serializable snapshot of build metadata.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Get returns the current build information.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a human-readable single-line version string.
func (i Info) String() string {
	return fmt.Sprintf("redintel-sentinel %s (commit: %s, built: %s, %s, %s)",
		i.Version, i.Commit, i.BuildTime, i.GoVersion, i.Platform)
}
