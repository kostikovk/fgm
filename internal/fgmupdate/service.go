package fgmupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/kostikovk/fgm/internal/toolhint"
	"github.com/kostikovk/fgm/internal/versionutil"
)

const (
	defaultBaseURL    = "https://api.github.com"
	defaultTTL        = 24 * time.Hour
	defaultReqTimeout = 1500 * time.Millisecond
	cacheFileName     = "update-check.json"
)

type cacheEntry struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
}

type githubReleasePayload struct {
	TagName string `json:"tag_name"`
}

// Config configures the FGM update checker service.
type Config struct {
	Root           string
	CurrentVersion string
	ExecutablePath string
	Client         *http.Client
	BaseURL        string
	Now            func() time.Time
	TTL            time.Duration
}

// Service checks and caches the latest available FGM release version.
type Service struct {
	root           string
	currentVersion string
	executablePath string
	client         *http.Client
	baseURL        string
	now            func() time.Time
	ttl            time.Duration
}

// New constructs an FGM update checker Service.
func New(cfg Config) *Service {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = defaultTTL
	}
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: defaultReqTimeout}
	}
	return &Service{
		root:           cfg.Root,
		currentVersion: normalizeVersion(cfg.CurrentVersion),
		executablePath: cfg.ExecutablePath,
		client:         client,
		baseURL:        baseURL,
		now:            nowFn,
		ttl:            ttl,
	}
}

// CheckLatest returns the latest release status, refreshing the local 24h cache from GitHub when expired.
func (s *Service) CheckLatest(ctx context.Context) (app.FGMUpdateStatus, bool, error) {
	if !isCheckableVersion(s.currentVersion) {
		return app.FGMUpdateStatus{}, false, nil
	}

	if entry, ok := s.readCache(); ok && s.now().Sub(entry.CheckedAt) < s.ttl {
		return s.buildStatus(entry.LatestVersion), true, nil
	}

	latest, err := s.fetchLatestRelease(ctx)
	if err != nil {
		if entry, ok := s.readCache(); ok && entry.LatestVersion != "" {
			return s.buildStatus(entry.LatestVersion), true, nil
		}
		return app.FGMUpdateStatus{}, false, err
	}

	_ = s.writeCache(cacheEntry{
		CheckedAt:     s.now().UTC(),
		LatestVersion: latest,
	})

	return s.buildStatus(latest), true, nil
}

// CachedNotice returns a cached update status from disk without performing any network I/O.
func (s *Service) CachedNotice(ctx context.Context) (app.FGMUpdateStatus, bool) {
	_ = ctx

	if !isCheckableVersion(s.currentVersion) {
		return app.FGMUpdateStatus{}, false
	}

	entry, ok := s.readCache()
	if !ok || entry.LatestVersion == "" {
		return app.FGMUpdateStatus{}, false
	}

	status := s.buildStatus(entry.LatestVersion)
	if !status.UpdateAvailable {
		return app.FGMUpdateStatus{}, false
	}
	return status, true
}

func (s *Service) buildStatus(latestVersion string) app.FGMUpdateStatus {
	cleanLatest := normalizeVersion(latestVersion)
	updateAvailable := cleanLatest != "" && versionutil.CompareVersions(cleanLatest, s.currentVersion) > 0
	return app.FGMUpdateStatus{
		CurrentVersion:  s.currentVersion,
		LatestVersion:   cleanLatest,
		UpdateAvailable: updateAvailable,
		UpgradeCommand:  toolhint.UpgradeFGMCommand(s.executablePath),
	}
}

func (s *Service) fetchLatestRelease(ctx context.Context) (string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, defaultReqTimeout)
	defer cancel()

	endpoint := s.baseURL + "/repos/kostikovk/fgm/releases/latest"
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("query latest fgm release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("latest fgm release returned status %d", resp.StatusCode)
	}

	var payload githubReleasePayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode latest fgm release: %w", err)
	}

	clean := normalizeVersion(payload.TagName)
	if !isCheckableVersion(clean) {
		return "", fmt.Errorf("invalid release tag %q", payload.TagName)
	}
	return clean, nil
}

func (s *Service) cachePath() string {
	return filepath.Join(s.root, "state", cacheFileName)
}

func (s *Service) readCache() (cacheEntry, bool) {
	if s.root == "" {
		return cacheEntry{}, false
	}
	data, err := os.ReadFile(s.cachePath())
	if err != nil {
		return cacheEntry{}, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return cacheEntry{}, false
	}
	entry.LatestVersion = normalizeVersion(entry.LatestVersion)
	if entry.LatestVersion == "" {
		return cacheEntry{}, false
	}
	return entry, true
}

func (s *Service) writeCache(entry cacheEntry) error {
	if s.root == "" {
		return nil
	}
	path := s.cachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

func isCheckableVersion(v string) bool {
	return v != "" && v != "dev" && v != "(devel)" && v != "unknown"
}
