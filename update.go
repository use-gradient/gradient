package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const ghReleasesAPI = "https://api.github.com/repos/use-gradient/gradient/releases/latest"

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func checkForUpdate() (latest string, downloadURL string, hasUpdate bool) {
	resp, err := http.Get(ghReleasesAPI)
	if err != nil || resp.StatusCode != 200 {
		return "", "", false
	}
	defer resp.Body.Close()
	var rel ghRelease
	if json.NewDecoder(resp.Body).Decode(&rel) != nil {
		return "", "", false
	}
	latest = strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(Version, "v")
	if latest == current || latest == "" {
		return latest, "", false
	}
	want := fmt.Sprintf("gradient_%s_%s", runtime.GOOS, runtime.GOARCH)
	for _, a := range rel.Assets {
		if a.Name == want {
			return latest, a.BrowserDownloadURL, true
		}
	}
	return latest, "", true
}

func runUpdate(args []string) int {
	fmt.Fprintf(os.Stderr, "Current version: %s\n", Version)
	fmt.Fprintln(os.Stderr, "Checking for updates...")

	latest, dlURL, hasUpdate := checkForUpdate()
	if !hasUpdate {
		fmt.Fprintln(os.Stderr, "Already up to date.")
		return 0
	}
	if dlURL == "" {
		fmt.Fprintf(os.Stderr, "New version %s available but no binary found for %s/%s.\n", latest, runtime.GOOS, runtime.GOARCH)
		fmt.Fprintln(os.Stderr, "Download manually from https://github.com/use-gradient/gradient/releases/latest")
		return 1
	}

	fmt.Fprintf(os.Stderr, "Downloading v%s...\n", latest)
	resp, err := http.Get(dlURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Error: download returned %s\n", resp.Status)
		return 1
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding current binary: %v\n", err)
		return 1
	}

	tmpPath := exe + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if _, err := f.ReadFrom(resp.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Error writing binary: %v\n", err)
		return 1
	}
	f.Close()

	if err := os.Rename(tmpPath, exe); err != nil {
		os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, "Error replacing binary: %v\nTry running with sudo.\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Updated to v%s\n", latest)
	if p := updateCachePath(); p != "" {
		os.Remove(p)
	}
	return 0
}

func hintUpdateIfAvailable() {
	if Version == "dev" {
		return
	}
	cachePath := updateCachePath()
	if cachePath == "" {
		return
	}
	info, err := os.Stat(cachePath)
	if err == nil && time.Since(info.ModTime()) < 24*time.Hour {
		b, _ := os.ReadFile(cachePath)
		msg := strings.TrimSpace(string(b))
		if msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
		return
	}
	latest, _, hasUpdate := checkForUpdate()
	var msg string
	if hasUpdate && latest != "" {
		msg = fmt.Sprintf("A new version of gradient is available (v%s). Run 'gradient update' to upgrade.", latest)
	}
	os.MkdirAll(filepath.Dir(cachePath), 0700)
	os.WriteFile(cachePath, []byte(msg), 0600)
	if msg != "" {
		fmt.Fprintln(os.Stderr, msg)
	}
}

func updateCachePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "gradient", "update_check")
}
