package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// UpdateRepo is the GitHub owner/repo whose latest release we ask about.
const UpdateRepo = "Jaace/breathe"

// UpdateCheckInterval is how long a successful check stays cached on disk
// before we'll hit the network again. 24h keeps us nice to GitHub's API
// rate limits while still surfacing new releases promptly.
const UpdateCheckInterval = 24 * time.Hour

// updateCacheFile is the JSON file we read/write to persist the most
// recent check result. Lives under $XDG_CACHE_HOME/breathe/ (falling back
// to ~/.cache/breathe/).
type updateCache struct {
	CheckedAt time.Time `json:"checked_at"`
	LatestTag string    `json:"latest_tag"`
}

// CheckLatestVersion returns the latest release tag for breathe (with the
// leading "v" stripped) if one is available and is strictly newer than
// `current`. Returns an empty string when the check is skipped, when no
// newer version exists, or when anything goes wrong — every error path is
// silent because a Pomodoro timer must always work without a network.
//
// Cached results younger than UpdateCheckInterval are reused without
// hitting the network.
func CheckLatestVersion(current string) string {
	if current == "" || current == "dev" {
		// Don't nag during local dev builds.
		return ""
	}

	cache, _ := readUpdateCache()
	now := time.Now()
	if cache != nil && now.Sub(cache.CheckedAt) < UpdateCheckInterval {
		return newerOrEmpty(current, cache.LatestTag)
	}

	latest := fetchLatestTag(3 * time.Second)
	if latest == "" {
		return ""
	}

	_ = writeUpdateCache(updateCache{CheckedAt: now, LatestTag: latest})
	return newerOrEmpty(current, latest)
}

func fetchLatestTag(timeout time.Duration) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", UpdateRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "breathe-update-check")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.TagName)
}

func updateCachePath() string {
	dir := os.Getenv("XDG_CACHE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".cache")
	}
	return filepath.Join(dir, "breathe", "update-check.json")
}

func readUpdateCache() (*updateCache, error) {
	path := updateCachePath()
	if path == "" {
		return nil, fmt.Errorf("no cache path")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c updateCache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func writeUpdateCache(c updateCache) error {
	path := updateCachePath()
	if path == "" {
		return fmt.Errorf("no cache path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// newerOrEmpty returns `latest` (with any leading "v" stripped) iff it is
// strictly newer than `current` under simple semver comparison. Returns
// "" when latest is missing, equal, or older — and when either tag fails
// to parse (we'd rather show nothing than something misleading).
func newerOrEmpty(current, latest string) string {
	cv := parseSemver(current)
	lv := parseSemver(latest)
	if cv == nil || lv == nil {
		return ""
	}
	if compareSemver(lv, cv) <= 0 {
		return ""
	}
	return strings.TrimPrefix(latest, "v")
}

// parseSemver pulls a [major, minor, patch] tuple out of a tag string,
// tolerating a leading "v" and ignoring any -prerelease / +build suffix.
// Returns nil if the tag isn't recognizable.
func parseSemver(tag string) []int {
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "v")
	if tag == "" {
		return nil
	}
	// Strip prerelease/build suffix.
	if i := strings.IndexAny(tag, "-+"); i >= 0 {
		tag = tag[:i]
	}
	parts := strings.Split(tag, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return nil
	}
	out := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil
		}
		out[i] = n
	}
	return out
}

func compareSemver(a, b []int) int {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			if a[i] > b[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}
