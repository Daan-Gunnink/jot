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
	"time"

	"github.com/fynelabs/selfupdate"
	"github.com/google/go-github/v60/github"
	"github.com/hashicorp/go-version"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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
	
	// Configure selfupdate logging
	selfupdate.LogInfo = func(format string, v ...interface{}) {
		fmt.Printf("[UPDATER INFO] "+format+"\n", v...)
	}
	selfupdate.LogError = func(format string, v ...interface{}) {
		fmt.Printf("[UPDATER ERROR] "+format+"\n", v...)
	}
}

// isBinaryAsset determines if the asset is a direct binary replacement
func isBinaryAsset(assetName string) bool {
	// Normalize the asset name to lowercase for comparison
	lowerName := strings.ToLower(assetName)
	osName := getOSName()
	
	// Define naming convention for binary assets
	switch osName {
	case "darwin":
		return strings.Contains(lowerName, "darwin") && 
			   strings.Contains(lowerName, "universal") &&
			   !strings.HasSuffix(lowerName, ".dmg") && 
			   !strings.HasSuffix(lowerName, ".pkg")
	case "windows":
		return strings.Contains(lowerName, "windows") && 
			   strings.Contains(lowerName, "amd64") && 
			   strings.HasSuffix(lowerName, ".exe") &&
			   !strings.Contains(lowerName, "installer")
	case "linux":
		return strings.Contains(lowerName, "linux") && 
			   !strings.HasSuffix(lowerName, ".deb") && 
			   !strings.HasSuffix(lowerName, ".rpm") &&
			   !strings.HasSuffix(lowerName, ".AppImage")
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

// applyBinaryUpdate applies a direct binary update using selfupdate
func applyBinaryUpdate(downloadPath string, ctx context.Context) error {
	// Notify user about the restart
	wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.InfoDialog,
		Title:   "Update Ready",
		Message: "The application will now update and restart automatically.",
	})
	
	// Create a local copy of the context to avoid issues in the goroutine
	localCtx := ctx
	
	// Apply the binary update and restart in a goroutine
	go func() {
		// Give the UI a moment to show the message dialog
		time.Sleep(1 * time.Second)
		
		// Open the downloaded binary inside the goroutine
		file, err := os.Open(downloadPath)
		if err != nil {
			fmt.Printf("Error opening update file: %v\n", err)
			wailsRuntime.MessageDialog(localCtx, wailsRuntime.MessageDialogOptions{
				Type:    wailsRuntime.ErrorDialog,
				Title:   "Update Failed",
				Message: fmt.Sprintf("Failed to open update file: %v", err),
			})
			return
		}
		// Close the file when done - within the goroutine
		defer file.Close()
		
		// Apply the update
		err = selfupdate.Apply(file, selfupdate.Options{})
		
		if err != nil {
			fmt.Printf("Error applying update: %v\n", err)
			wailsRuntime.MessageDialog(localCtx, wailsRuntime.MessageDialogOptions{
				Type:    wailsRuntime.ErrorDialog,
				Title:   "Update Failed",
				Message: fmt.Sprintf("Failed to apply update: %v", err),
			})
			return
		}
		
		// Try to restart the application after update
		restartPath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting executable path: %v\n", err)
		} else {
			// Start the updated application
			cmd := exec.Command(restartPath)
			err = cmd.Start()
			if err != nil {
				fmt.Printf("Error restarting: %v\n", err)
			}
		}
		
		// Quit the application anyway, as it has been updated
		wailsRuntime.Quit(localCtx)
	}()
	
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
		
		// Create a local copy of the context to avoid issues in the goroutine
		localCtx := ctx
		
		// Run in a goroutine so we don't block the UI
		go func() {
			// Open the DMG
			cmd := exec.Command("open", downloadPath)
			err := cmd.Run()
			if err != nil {
				fmt.Printf("Error opening DMG: %v\n", err)
				return
			}
			
			// Wait a bit to let the user install
			time.Sleep(5 * time.Second)
			
			// Show message about quitting
			wailsRuntime.MessageDialog(localCtx, wailsRuntime.MessageDialogOptions{
				Type:    wailsRuntime.InfoDialog,
				Title:   "Update Complete",
				Message: "The application will now quit. Please restart it after installation.",
			})
			
			// Wait a bit for user to read the message
			time.Sleep(2 * time.Second)
			
			// Quit the app
			wailsRuntime.Quit(localCtx)
		}()
		
		return nil
	} else if strings.HasSuffix(downloadPath, ".pkg") {
		// Similar process for PKG files
		wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.InfoDialog,
			Title:   "Update Instructions",
			Message: "The update will now install. Please follow the installation instructions and restart the application.",
		})
		
		// Create a local copy of the context to avoid issues in the goroutine
		localCtx := ctx
		
		go func() {
			cmd := exec.Command("open", downloadPath)
			err := cmd.Run()
			if err != nil {
				fmt.Printf("Error opening PKG: %v\n", err)
				return
			}
			
			// Wait a bit to let the user install
			time.Sleep(5 * time.Second)
			
			// Show message about quitting
			wailsRuntime.MessageDialog(localCtx, wailsRuntime.MessageDialogOptions{
				Type:    wailsRuntime.InfoDialog,
				Title:   "Update Complete",
				Message: "The application will now quit. Please restart it after installation.",
			})
			
			// Wait a bit for user to read the message
			time.Sleep(2 * time.Second)
			
			// Quit the app
			wailsRuntime.Quit(localCtx)
		}()
		
		return nil
	}
	
	return fmt.Errorf("unsupported file format for macOS: %s", downloadPath)
}

func applyWindowsUpdate(downloadPath string, ctx context.Context) error {
	// For Windows, we typically work with .exe or .msi files
	wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.InfoDialog,
		Title:   "Update Instructions",
		Message: "The update will now install. Please follow the installation instructions. The application will restart automatically after installation.",
	})
	
	// Create a local copy of the context to avoid issues in the goroutine
	localCtx := ctx
	
	// Run the installer in a goroutine
	go func() {
		var installErr error
		
		// Start the installer
		if strings.HasSuffix(strings.ToLower(downloadPath), ".msi") {
			// For MSI installers
			cmd := exec.Command("msiexec", "/i", downloadPath, "/quiet")
			installErr = cmd.Run()
		} else {
			// For EXE installers
			cmd := exec.Command(downloadPath)
			installErr = cmd.Run()
		}
		
		if installErr != nil {
			fmt.Printf("Error running installer: %v\n", installErr)
			wailsRuntime.MessageDialog(localCtx, wailsRuntime.MessageDialogOptions{
				Type:    wailsRuntime.ErrorDialog,
				Title:   "Update Failed",
				Message: fmt.Sprintf("Failed to run installer: %v", installErr),
			})
			return
		}
		
		// Wait for the installation to complete
		time.Sleep(10 * time.Second)
		
		// Show a message that we're restarting
		wailsRuntime.MessageDialog(localCtx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.InfoDialog,
			Title:   "Update Complete",
			Message: "The update has been installed. The application will now restart.",
		})
		
		// Wait a bit for user to read the message
		time.Sleep(2 * time.Second)
		
		// Try to restart the application
		restartPath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting executable path: %v\n", err)
			wailsRuntime.MessageDialog(localCtx, wailsRuntime.MessageDialogOptions{
				Type:    wailsRuntime.ErrorDialog,
				Title:   "Restart Failed",
				Message: "Could not restart the application automatically. Please restart it manually.",
			})
		} else {
			// Start the updated application - create a detached process in Windows
			cmd := exec.Command("cmd", "/C", "start", restartPath)
			err = cmd.Start()
			if err != nil {
				fmt.Printf("Error restarting: %v\n", err)
				wailsRuntime.MessageDialog(localCtx, wailsRuntime.MessageDialogOptions{
					Type:    wailsRuntime.ErrorDialog,
					Title:   "Restart Failed",
					Message: "Could not restart the application automatically. Please restart it manually.",
				})
			}
		}
		
		// Quit this instance
		wailsRuntime.Quit(localCtx)
	}()
	
	return nil
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