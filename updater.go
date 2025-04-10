package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/hashicorp/go-version"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetAppVersion returns the current version of the application.
// It first tries to read from APP_VERSION environment variable,
// and falls back to "0.1.0" if not set.
func GetAppVersion() string {
	version := os.Getenv("APP_VERSION")
	fmt.Println("version", version)
	if version == "" {
		version = "0.0.0" // Default fallback version
	}
	return version
}

// CurrentVersion is the current version of the application.
// This should be updated with each new release.
const CurrentVersion = "0.1.0"

// GitHubInfo contains the information needed to check for updates.
type GitHubInfo struct {
	Owner string
	Repo  string
}

// UpdaterService handles checking for and applying updates.
type UpdaterService struct {
	ctx       context.Context
	githubInfo GitHubInfo
}

// NewUpdaterService creates a new updater service.
func NewUpdaterService(owner, repo string) *UpdaterService {
	return &UpdaterService{
		githubInfo: GitHubInfo{
			Owner: owner,
			Repo:  repo,
		},
	}
}

// Initialize stores the context for later use.
func (u *UpdaterService) Initialize(ctx context.Context) {
	u.ctx = ctx
}

// CheckForUpdates checks if a newer version is available.
func (u *UpdaterService) CheckForUpdates() (bool, string, error) {
	client := github.NewClient(nil)
	
	// Get the latest release from GitHub
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), u.githubInfo.Owner, u.githubInfo.Repo)
	if err != nil {
		return false, "", fmt.Errorf("error checking for updates: %w", err)
	}
	
	// Compare versions
	latestVersion := strings.TrimPrefix(*release.TagName, "v")
	fmt.Println("latestVersion", latestVersion)
	
	currentVersion := GetAppVersion()
	fmt.Println("currentVersion", currentVersion)
	currentV, err := version.NewVersion(currentVersion)
	if err != nil {
		return false, "", fmt.Errorf("error parsing current version: %w", err)
	}
	
	latestV, err := version.NewVersion(latestVersion)
	if err != nil {
		return false, "", fmt.Errorf("error parsing latest version: %w", err)
	}
	
	// Return whether an update is available and the latest version
	if latestV.GreaterThan(currentV) {
		return true, latestVersion, nil
	}
	
	return false, latestVersion, nil
}

// DownloadUpdate downloads the latest release for the current platform.
func (u *UpdaterService) DownloadUpdate() (string, error) {
	client := github.NewClient(nil)
	
	// Get the latest release
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), u.githubInfo.Owner, u.githubInfo.Repo)
	if err != nil {
		return "", fmt.Errorf("error getting latest release: %w", err)
	}
	
	// Determine which asset to download based on the current OS
	var assetURL string
	var assetName string
	
	for _, asset := range release.Assets {
		name := *asset.Name
		if matchesPlatform(name) {
			assetURL = *asset.BrowserDownloadURL
			assetName = name
			break
		}
	}
	
	if assetURL == "" {
		return "", fmt.Errorf("no suitable download found for this platform")
	}
	
	// Create temp directory for download
	tempDir, err := os.MkdirTemp("", "jot-update")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %w", err)
	}
	
	// Download the file
	downloadPath := filepath.Join(tempDir, assetName)
	err = downloadFile(assetURL, downloadPath)
	if err != nil {
		return "", fmt.Errorf("error downloading update: %w", err)
	}
	
	return downloadPath, nil
}

// ApplyUpdate applies the downloaded update.
// Note: The implementation details vary greatly depending on platform
func (u *UpdaterService) ApplyUpdate(downloadPath string) error {
	// Show dialog to confirm update installation
	selection, err := wailsRuntime.MessageDialog(u.ctx, wailsRuntime.MessageDialogOptions{
		Type:          wailsRuntime.QuestionDialog,
		Title:         "Update Ready",
		Message:       "An update has been downloaded. The application will restart to install it. Continue?",
		Buttons:       []string{"Install", "Cancel"},
		DefaultButton: "Install",
		CancelButton:  "Cancel",
	})
	
	if err != nil {
		return fmt.Errorf("error showing dialog: %w", err)
	}
	
	if selection == "Cancel" {
		return nil
	}
	
	// Execute the update based on platform
	osName := getOSName()
	
	switch osName {
	case "darwin":
		// For macOS
		return applyMacOSUpdate(downloadPath, u.ctx)
	case "windows":
		// For Windows
		return applyWindowsUpdate(downloadPath, u.ctx)
	case "linux":
		// For Linux
		return applyLinuxUpdate(downloadPath, u.ctx)
	default:
		return fmt.Errorf("unsupported platform: %s", osName)
	}
}

// Helper functions

// getOSName returns the current OS name
func getOSName() string {
	return runtime.GOOS
}

// matchesPlatform checks if the asset name matches the current platform
func matchesPlatform(assetName string) bool {
	assetName = strings.ToLower(assetName)
	osName := getOSName()
	
	switch osName {
	case "windows":
		return strings.Contains(assetName, "windows") || strings.HasSuffix(assetName, ".exe") || strings.HasSuffix(assetName, ".msi")
	case "darwin":
		return strings.Contains(assetName, "darwin") || strings.Contains(assetName, "mac") || strings.HasSuffix(assetName, ".dmg") || strings.HasSuffix(assetName, ".pkg")
	case "linux":
		return strings.Contains(assetName, "linux") || strings.HasSuffix(assetName, ".deb") || strings.HasSuffix(assetName, ".rpm") || strings.HasSuffix(assetName, ".AppImage")
	default:
		return false
	}
}

// downloadFile downloads a file from a URL to a local path
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	
	_, err = io.Copy(out, resp.Body)
	return err
}

// Platform-specific update applications

func applyMacOSUpdate(downloadPath string, ctx context.Context) error {
	// For macOS, we typically work with .dmg or .pkg files
	if strings.HasSuffix(downloadPath, ".dmg") {
		// Mount the DMG
		wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.InfoDialog,
			Title:   "Update Instructions",
			Message: "The update will now open. Please follow the installation instructions and restart the application.",
		})
		
		cmd := exec.Command("open", downloadPath)
		return cmd.Run()
	} else if strings.HasSuffix(downloadPath, ".pkg") {
		wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.InfoDialog,
			Title:   "Update Instructions",
			Message: "The update will now install. Please follow the installation instructions and restart the application.",
		})
		
		cmd := exec.Command("open", downloadPath)
		return cmd.Run()
	}
	
	return fmt.Errorf("unsupported file format for macOS: %s", downloadPath)
}

func applyWindowsUpdate(downloadPath string, ctx context.Context) error {
	// For Windows, we typically work with .exe or .msi files
	wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.InfoDialog,
		Title:   "Update Instructions",
		Message: "The update will now install. Please follow the installation instructions and restart the application.",
	})
	
	cmd := exec.Command("cmd", "/C", "start", downloadPath)
	return cmd.Run()
}

func applyLinuxUpdate(downloadPath string, ctx context.Context) error {
	// For Linux, we might have different package formats
	if strings.HasSuffix(downloadPath, ".deb") {
		wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.InfoDialog,
			Title:   "Update Instructions",
			Message: "The update will now install. You may be prompted for your password.",
		})
		
		cmd := exec.Command("xdg-open", downloadPath)
		return cmd.Run()
	} else if strings.HasSuffix(downloadPath, ".rpm") {
		wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.InfoDialog,
			Title:   "Update Instructions",
			Message: "The update will now install. You may be prompted for your password.",
		})
		
		cmd := exec.Command("xdg-open", downloadPath)
		return cmd.Run()
	} else if strings.HasSuffix(downloadPath, ".AppImage") {
		// For AppImage, make it executable and run it
		wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.InfoDialog,
			Title:   "Update Instructions",
			Message: "The update has been downloaded. The application will close and you can run the new version.",
		})
		
		os.Chmod(downloadPath, 0755)
		cmd := exec.Command("xdg-open", filepath.Dir(downloadPath))
		return cmd.Run()
	}
	
	return fmt.Errorf("unsupported file format for Linux: %s", downloadPath)
} 