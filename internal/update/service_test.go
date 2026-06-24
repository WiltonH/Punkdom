package update

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSelectAssetForPlatform(t *testing.T) {
	assets := []githubAsset{
		{Name: "checksums.txt"},
		{Name: "punkdom-v0.1.11-darwin-arm64.tar.gz", DownloadURL: "asset-api-url"},
		{Name: "punkdom-v0.1.11-linux-x64.tar.gz"},
	}
	asset := selectAsset(assets, "darwin-arm64")
	if asset == nil || asset.Name != "punkdom-v0.1.11-darwin-arm64.tar.gz" {
		t.Fatalf("unexpected asset: %#v", asset)
	}
	if got := selectAsset(assets, "windows-x64"); got != nil {
		t.Fatalf("windows asset should not match: %#v", got)
	}
}

func TestPlatformKeyNormalizesAMD64(t *testing.T) {
	if got := platformKey("darwin", "amd64"); got != "darwin-x64" {
		t.Fatalf("platformKey darwin/amd64 = %s", got)
	}
	if got := platformKey("linux", "arm64"); got != "linux-arm64" {
		t.Fatalf("platformKey linux/arm64 = %s", got)
	}
}

func TestCheckFallsBackToReleaseRedirectWhenGitHubAPIRateLimited(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/repos/WiltonH/Punkdom/releases/latest":
			return testResponse(req, http.StatusForbidden, "", `{"message":"API rate limit exceeded"}`), nil
		case "/WiltonH/Punkdom/releases/latest":
			return testResponse(req, http.StatusFound, "/WiltonH/Punkdom/releases/tag/v0.1.1", ""), nil
		default:
			return testResponse(req, http.StatusNotFound, "", "not found"), nil
		}
	})}

	service := &Service{
		repository:     "WiltonH/Punkdom",
		currentVersion: "0.1.1",
		httpClient:     client,
		apiBaseURL:     "https://example.test/repos/",
		webBaseURL:     "https://example.test/",
	}
	result, err := service.Check(context.Background())
	if err != nil {
		t.Fatalf("Check should fall back without error: %v", err)
	}
	if result.LatestVersion != "0.1.1" {
		t.Fatalf("LatestVersion = %q, want 0.1.1", result.LatestVersion)
	}
	if result.UpdateAvailable {
		t.Fatalf("UpdateAvailable should be false for same fallback version")
	}
	if result.ReleaseURL != "https://example.test/WiltonH/Punkdom/releases/tag/v0.1.1" {
		t.Fatalf("ReleaseURL = %q", result.ReleaseURL)
	}
}

func TestCheckDisablesInstallInDockerRuntime(t *testing.T) {
	t.Setenv("PUNKDOM_DOCKER", "1")
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return testResponse(req, http.StatusOK, "", `{
			"tag_name":"v0.1.2",
			"html_url":"https://example.test/WiltonH/Punkdom/releases/tag/v0.1.2",
			"assets":[{"name":"punkdom-v0.1.2-linux-x64.tar.gz","url":"asset-api-url","browser_download_url":"asset-web-url","size":123}]
		}`), nil
	})}

	service := &Service{
		repository:     "WiltonH/Punkdom",
		currentVersion: "0.1.1",
		httpClient:     client,
		apiBaseURL:     "https://example.test/repos/",
		webBaseURL:     "https://example.test/",
	}
	result, err := service.Check(context.Background())
	if err != nil {
		t.Fatalf("Check should work in Docker runtime: %v", err)
	}
	if !result.Docker {
		t.Fatalf("Docker flag should be true")
	}
	if !result.UpdateAvailable {
		t.Fatalf("UpdateAvailable should be true")
	}
	if result.CanInstall {
		t.Fatalf("Docker runtime should not allow application-level install")
	}
	if result.DockerCommand == "" || !strings.Contains(result.DockerCommand, "docker compose pull punkdom") {
		t.Fatalf("DockerCommand missing update command: %q", result.DockerCommand)
	}
}

func TestInstallRejectsDockerRuntime(t *testing.T) {
	t.Setenv("PUNKDOM_DOCKER", "true")
	service := &Service{}
	if _, err := service.Install(context.Background()); err == nil || !strings.Contains(err.Error(), "Docker") {
		t.Fatalf("Install should reject Docker runtime, got %v", err)
	}
}

func TestReleaseTagFromURL(t *testing.T) {
	tests := []string{
		"https://github.com/WiltonH/Punkdom/releases/tag/v0.1.1",
		"/WiltonH/Punkdom/releases/tag/v0.1.1",
	}
	for _, rawURL := range tests {
		if got, err := releaseTagFromURL("WiltonH/Punkdom", rawURL); err != nil || got != "v0.1.1" {
			t.Fatalf("releaseTagFromURL(%q) = %q, %v", rawURL, got, err)
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testResponse(req *http.Request, status int, location, body string) *http.Response {
	header := http.Header{}
	if location != "" {
		header.Set("Location", location)
	}
	return &http.Response{
		StatusCode: status,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}
