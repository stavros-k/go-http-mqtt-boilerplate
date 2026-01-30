package utils

import (
	"strings"
	"testing"
)

func TestGetBuildVersion(t *testing.T) {
	t.Parallel()

	version := GetBuildVersion()

	if version == "" {
		t.Error("GetBuildVersion() should not return empty string")
	}

	// Should contain version format
	if !strings.Contains(version, "v") {
		t.Errorf("GetBuildVersion() should contain 'v' prefix, got: %s", version)
	}

	// Should contain commit hash in parentheses
	if !strings.Contains(version, "(") || !strings.Contains(version, ")") {
		t.Errorf("GetBuildVersion() should contain commit hash in parentheses, got: %s", version)
	}

	// Should contain "built at"
	if !strings.Contains(version, "built at") {
		t.Errorf("GetBuildVersion() should contain 'built at', got: %s", version)
	}
}

func TestGetVersionShort(t *testing.T) {
	t.Parallel()

	version := GetVersionShort()

	if version == "" {
		t.Error("GetVersionShort() should not return empty string")
	}

	// Should contain version format
	if !strings.Contains(version, "v") {
		t.Errorf("GetVersionShort() should contain 'v' prefix, got: %s", version)
	}

	// Should contain commit hash in parentheses
	if !strings.Contains(version, "(") || !strings.Contains(version, ")") {
		t.Errorf("GetVersionShort() should contain commit hash in parentheses, got: %s", version)
	}

	// Should NOT contain "built at" (short version)
	if strings.Contains(version, "built at") {
		t.Errorf("GetVersionShort() should not contain 'built at', got: %s", version)
	}
}

func TestGetBuildInfo(t *testing.T) {
	t.Parallel()

	info := GetBuildInfo()

	if info == nil {
		t.Fatal("GetBuildInfo() should not return nil")
	}

	requiredKeys := []string{"version", "commit", "build_time", "vcs_modified"}
	for _, key := range requiredKeys {
		if _, ok := info[key]; !ok {
			t.Errorf("GetBuildInfo() should contain key %q", key)
		}
	}

	// Version should not be empty
	if info["version"] == "" {
		t.Error("GetBuildInfo()['version'] should not be empty")
	}

	// vcs_modified should be "true" or "false"
	modified := info["vcs_modified"]
	if modified != "true" && modified != "false" {
		t.Errorf("GetBuildInfo()['vcs_modified'] should be 'true' or 'false', got: %s", modified)
	}

	// Should contain go_version if available
	if goVersion, ok := info["go_version"]; ok {
		if !strings.HasPrefix(goVersion, "go") {
			t.Errorf("GetBuildInfo()['go_version'] should start with 'go', got: %s", goVersion)
		}
	}
}

func TestGetVCSInfo(t *testing.T) {
	t.Parallel()

	commit, buildTime, modified := getVCSInfo()

	// Commit should not be empty
	if commit == "" {
		t.Error("getVCSInfo() commit should not be empty")
	}

	// Build time should not be empty
	if buildTime == "" {
		t.Error("getVCSInfo() buildTime should not be empty")
	}

	// Modified should be "true" or "false"
	if modified != "true" && modified != "false" {
		t.Errorf("getVCSInfo() modified should be 'true' or 'false', got: %s", modified)
	}

	// If commit was shortened, it should be 7 characters or less (unless it's "unknown")
	if commit != "unknown" && len(commit) > 7 {
		t.Errorf("getVCSInfo() commit should be shortened to 7 chars, got: %s (len=%d)", commit, len(commit))
	}
}

func TestBuildVersionContainsModifiedFlag(t *testing.T) {
	t.Parallel()

	// This test checks that the -dirty suffix is properly handled
	_, _, modified := getVCSInfo()

	version := GetBuildVersion()
	versionShort := GetVersionShort()

	if modified == "true" {
		if !strings.Contains(version, "-dirty") {
			t.Error("GetBuildVersion() should contain '-dirty' when vcs.modified is true")
		}
		if !strings.Contains(versionShort, "-dirty") {
			t.Error("GetVersionShort() should contain '-dirty' when vcs.modified is true")
		}
	} else {
		// Can't assert absence of -dirty since it might be in the version string for other reasons
		// Just verify the functions run without error
		if version == "" || versionShort == "" {
			t.Error("Version functions should return non-empty strings")
		}
	}
}

func TestVersionConsistency(t *testing.T) {
	t.Parallel()

	// Build info version should match the version variable
	info := GetBuildInfo()
	if info["version"] != Version {
		t.Errorf("GetBuildInfo()['version'] = %s, want %s", info["version"], Version)
	}

	// Both version functions should use the same version string
	version := GetBuildVersion()
	versionShort := GetVersionShort()

	if !strings.Contains(version, "v"+Version) {
		t.Errorf("GetBuildVersion() should contain version %s, got: %s", Version, version)
	}

	if !strings.Contains(versionShort, "v"+Version) {
		t.Errorf("GetVersionShort() should contain version %s, got: %s", Version, versionShort)
	}
}

func TestCommitHashFormat(t *testing.T) {
	t.Parallel()

	commit, _, _ := getVCSInfo()

	// If commit is not "unknown", verify it looks like a git hash
	if commit != "unknown" && commit != "" {
		// Git short hash should be 7 characters or less (hex characters)
		for _, c := range commit {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("Commit hash should be hexadecimal, got: %s", commit)
				break
			}
		}
	}
}