// Package version exposes build metadata, injected at build time via -ldflags.
package version

var (
	// Version is the portfolio release, e.g. a semver or git short SHA.
	Version = "dev"
	// BuildTime is an RFC3339 timestamp of when the binary was built.
	BuildTime = ""
)
