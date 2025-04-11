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
	"github.com/inconshreveable/go-update"
)

// Version is the current version of the application.
// This will be overridden during build time using ldflags
var Version = "0.0.0"

// GetAppVersion returns the current version of the application.
// It uses the Version variable that's set during compilation
func GetAppVersion() string {
	// First check environment variable for development/testing
	envVersion := os.Getenv("APP_VERSION")
	if envVersion != "" {
		return envVersion
	}
	
	// Return the version set at build time
	return Version
}

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

// UpdateType defines the type of update available
type UpdateType int

const (
	// BinaryUpdate is a direct binary replacement using go-update
	BinaryUpdate UpdateType = iota
	// PackageUpdate is a packaged update (.dmg, .exe, etc.)
	PackageUpdate
)

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Type        UpdateType
	Version     string
	DownloadURL string
	AssetName   string
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

// isBinaryAsset determines if the asset is a direct binary replacement
func isBinaryAsset(assetName string) bool {
	// Binary assets should be named consistently
	// For example: jot-darwin-amd64, jot-windows-amd64.exe, jot-linux-amd64
	osName := getOSName()
	
	// Define naming convention for binary assets
	switch osName {
	case "darwin":
		return strings.Contains(assetName, "darwin") && 
			   !strings.HasSuffix(assetName, ".dmg") && 
			   !strings.HasSuffix(assetName, ".pkg")
	case "windows":
		return strings.Contains(assetName, "windows") && 
			   strings.HasSuffix(assetName, ".exe") &&
			   !strings.Contains(assetName, "installer")
	case "linux":
		return strings.Contains(assetName, "linux") && 
			   !strings.HasSuffix(assetName, ".deb") && 
			   !strings.HasSuffix(assetName, ".rpm") &&
			   !strings.HasSuffix(assetName, ".AppImage")
	default:
		return false
	}
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
	
	// Check if there are assets available for the current platform
	platformAssetAvailable := false
	for _, asset := range release.Assets {
		if asset.Name != nil && matchesPlatform(*asset.Name) {
			platformAssetAvailable = true
			break
		}
	}
	
	// Return whether an update is available, the latest version, and if it's available for this platform
	if latestV.GreaterThan(currentV) && platformAssetAvailable {
		return true, latestVersion, nil
	}
	
	return false, latestVersion, nil
}

// GetUpdateInfo retrieves detailed information about the available update
func (u *UpdaterService) GetUpdateInfo() (*UpdateInfo, error) {
	client := github.NewClient(nil)
	
	// Get the latest release
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), u.githubInfo.Owner, u.githubInfo.Repo)
	if err != nil {
		return nil, fmt.Errorf("error getting latest release: %w", err)
	}
	
	// Find the appropriate asset for the current platform
	for _, asset := range release.Assets {
		name := *asset.Name
		if matchesPlatform(name) {
			updateType := PackageUpdate
			if isBinaryAsset(name) {
				updateType = BinaryUpdate
			}
			
			return &UpdateInfo{
				Type:        updateType,
				Version:     strings.TrimPrefix(*release.TagName, "v"),
				DownloadURL: *asset.BrowserDownloadURL,
				AssetName:   name,
			}, nil
		}
	}
	
	return nil, fmt.Errorf("no suitable update found for this platform")
}

// DownloadUpdate downloads the latest release for the current platform.
func (u *UpdaterService) DownloadUpdate() (string, *UpdateInfo, error) {
	// Get update info
	updateInfo, err := u.GetUpdateInfo()
	if err != nil {
		return "", nil, err
	}
	
	// Create temp directory for download
	tempDir, err := os.MkdirTemp("", "jot-update")
	if err != nil {
		return "", nil, fmt.Errorf("error creating temp directory: %w", err)
	}
	
	// Download the file
	downloadPath := filepath.Join(tempDir, updateInfo.AssetName)
	err = downloadFile(updateInfo.DownloadURL, downloadPath)
	if err != nil {
		return "", nil, fmt.Errorf("error downloading update: %w", err)
	}
	
	return downloadPath, updateInfo, nil
}

// ApplyUpdate applies the downloaded update.
func (u *UpdaterService) ApplyUpdate(downloadPath string, updateInfo *UpdateInfo) error {
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
	
	// Apply update based on the update type
	if updateInfo.Type == BinaryUpdate {
		return applyBinaryUpdate(downloadPath, u.ctx)
	} else {
		// Execute the platform-specific update for packaged updates
		osName := getOSName()
		
		switch osName {
		case "darwin":
			return applyMacOSUpdate(downloadPath, u.ctx)
		case "windows":
			return applyWindowsUpdate(downloadPath, u.ctx)
		case "linux":
			return applyLinuxUpdate(downloadPath, u.ctx)
		default:
			return fmt.Errorf("unsupported platform: %s", osName)
		}
	}
}

// applyBinaryUpdate applies a direct binary update using go-update
func applyBinaryUpdate(downloadPath string, ctx context.Context) error {
	// Notify user about the restart
	wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.InfoDialog,
		Title:   "Update Ready",
		Message: "The application will now update and restart automatically.",
	})
	
	// Open the downloaded binary
	file, err := os.Open(downloadPath)
	if err != nil {
		return fmt.Errorf("error opening update file: %w", err)
	}
	defer file.Close()
	
	// Apply the binary update
	err = update.Apply(file, update.Options{})
	if err != nil {
		return fmt.Errorf("error applying update: %w", err)
	}
	
	// Restart the application
	// This is platform specific and might require additional logic
	// The application will restart with the new binary once the current process exits
	
	// Note: For Wails applications, we might need to use the Wails runtime to quit
	// and have a separate mechanism to restart the app
	wailsRuntime.Quit(ctx)
	return nil
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