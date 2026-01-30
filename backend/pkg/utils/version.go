package utils

import (
	"fmt"
	"runtime/debug"
)

// Version information that can be set at build time using ldflags
//
//nolint:gochecknoglobals // Build-time variables set via -ldflags during compilation
var (
	// Version is the semantic version of the build.
	Version = "dev"
	// Commit is the git commit hash.
	Commit = "unknown"
	// BuildTime is when the binary was built.
	BuildTime = "unknown"
)

// getVCSInfo retrieves VCS information from runtime/debug if available.
func getVCSInfo() (commit string, buildTime string, modified string) {
	commit = Commit
	buildTime = BuildTime
	modified = "false"

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				if Commit == "unknown" || Commit == "" {
					commit = setting.Value
					if len(commit) > 7 {
						commit = commit[:7] // Short hash
					}
				}
			case "vcs.time":
				if BuildTime == "unknown" || BuildTime == "" {
					buildTime = setting.Value
				}
			case "vcs.modified":
				modified = setting.Value
			}
		}
	}

	return commit, buildTime, modified
}

// GetBuildVersion returns a formatted build version string
// It combines the version, commit hash, and build time into a single string
// Example output: "v1.0.0 (abc1234) built at 2024-01-15T10:30:00Z".
func GetBuildVersion() string {
	commit, buildTime, modified := getVCSInfo()

	suffix := ""
	if modified == "true" {
		suffix = "-dirty"
	}

	return fmt.Sprintf("v%s (%s%s) built at %s", Version, commit, suffix, buildTime)
}

func GetVersionShort() string {
	commit, _, modified := getVCSInfo()

	suffix := ""
	if modified == "true" {
		suffix = "-dirty"
	}

	return fmt.Sprintf("v%s (%s%s)", Version, commit, suffix)
}

// GetBuildInfo returns detailed build information including Go version and dependencies
// This uses runtime/debug to get VCS information if available (Go 1.18+).
func GetBuildInfo() map[string]string {
	commit, buildTime, modified := getVCSInfo()

	info := map[string]string{
		"version":      Version,
		"commit":       commit,
		"build_time":   buildTime,
		"vcs_modified": modified,
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		info["go_version"] = buildInfo.GoVersion
	}

	return info
}
