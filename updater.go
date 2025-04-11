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
	
	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}
	
	// Apply the binary update
	err = update.Apply(file, update.Options{})
	if err != nil {
		return fmt.Errorf("error applying update: %w", err)
	}
	
	// Restart the application
	go func() {
		// Give the UI a moment to show the message dialog
		time.Sleep(2 * time.Second)
		
		// Start the new version of the app
		restartApp(execPath)
		
		// Quit this instance
		wailsRuntime.Quit(ctx)
	}()
	
	return nil
}

// restartApp starts the application after update
func restartApp(execPath string) error {
	// Create the command to restart the app
	var cmd *exec.Cmd
	
	switch getOSName() {
	case "darwin":
		// For macOS, we need to find the .app container
		appBundle := execPath
		// If we're in the MacOS/Contents folder of an app bundle, go up to the .app
		if strings.Contains(execPath, "Contents/MacOS") {
			// Go up three levels from the binary to get to the .app
			// e.g. /Applications/MyApp.app/Contents/MacOS/myapp -> /Applications/MyApp.app
			appBundle = filepath.Dir(filepath.Dir(filepath.Dir(execPath)))
		}
		cmd = exec.Command("open", appBundle)
	case "windows":
		// For Windows, run the executable directly but detach from parent process
		cmd = exec.Command("cmd", "/C", "start", execPath)
	case "linux":
		// For Linux, run the executable directly
		cmd = exec.Command(execPath)
	}
	
	if cmd != nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Start()
	}
	
	return fmt.Errorf("unsupported platform for restart")
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
		
		// Get the path to the current executable
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error getting executable path: %w", err)
		}
		
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
			wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
				Type:    wailsRuntime.InfoDialog,
				Title:   "Update Complete",
				Message: "The application will now quit. Please restart it after installation.",
			})
			
			// Wait a bit for user to read the message
			time.Sleep(2 * time.Second)
			
			// Quit the app
			wailsRuntime.Quit(ctx)
		}()
		
		return nil
	} else if strings.HasSuffix(downloadPath, ".pkg") {
		// Similar process for PKG files
		wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.InfoDialog,
			Title:   "Update Instructions",
			Message: "The update will now install. Please follow the installation instructions and restart the application.",
		})
		
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
			wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
				Type:    wailsRuntime.InfoDialog,
				Title:   "Update Complete",
				Message: "The application will now quit. Please restart it after installation.",
			})
			
			// Wait a bit for user to read the message
			time.Sleep(2 * time.Second)
			
			// Quit the app
			wailsRuntime.Quit(ctx)
		}()
		
		return nil
	}
	
	return fmt.Errorf("unsupported file format for macOS: %s", downloadPath)
}

func applyWindowsUpdate(downloadPath string, ctx context.Context) error {
	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}
	
	// For Windows, we typically work with .exe or .msi files
	wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.InfoDialog,
		Title:   "Update Instructions",
		Message: "The update will now install. Please follow the installation instructions. The application will restart automatically after installation.",
	})
	
	// Run the installer in a goroutine
	go func() {
		// Start the installer
		if strings.HasSuffix(strings.ToLower(downloadPath), ".msi") {
			// For MSI installers
			cmd := exec.Command("msiexec", "/i", downloadPath, "/quiet")
			err := cmd.Run()
			if err != nil {
				fmt.Printf("Error running MSI installer: %v\n", err)
				return
			}
		} else {
			// For EXE installers
			cmd := exec.Command(downloadPath)
			err := cmd.Run()
			if err != nil {
				fmt.Printf("Error running EXE installer: %v\n", err)
				return
			}
		}
		
		// Wait for the installation to complete
		time.Sleep(10 * time.Second)
		
		// Show a message that we're restarting
		wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.InfoDialog,
			Title:   "Update Complete",
			Message: "The update has been installed. The application will now restart.",
		})
		
		// Wait a bit for user to read the message
		time.Sleep(2 * time.Second)
		
		// Restart the application
		restartApp(execPath)
		
		// Quit this instance
		wailsRuntime.Quit(ctx)
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