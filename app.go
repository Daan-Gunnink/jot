package main

import (
	"context"
	"fmt"
	"time"
)

// App struct
type App struct {
	ctx     context.Context
	updater *UpdaterService
}

// NewApp creates a new App application struct
func NewApp() *App {
	updater := NewUpdaterService("daan-gunnink", "toJot")
	return &App{
		updater: updater,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.updater.Initialize(ctx)
	
	// Check for updates on startup (after a short delay to let the UI load)
	go func() {
		time.Sleep(2 * time.Second)
		a.CheckForUpdates()
	}()
}

// CheckForUpdates checks if updates are available and prompts the user if they are
func (a *App) CheckForUpdates() string {
	hasUpdate, latestVersion, err := a.updater.CheckForUpdates()
	if err != nil {
		return fmt.Sprintf("Error checking for updates: %s", err.Error())
	}
	
	if hasUpdate {
		return fmt.Sprintf("Update available! Version %s is available (current: %s)", 
			latestVersion, Version)
	}
	
	return fmt.Sprintf("You're running the latest version: %s", Version)
}

// DownloadAndInstallUpdate downloads and installs the latest update
func (a *App) DownloadAndInstallUpdate() string {
	// Download the update
	downloadPath, updateInfo, err := a.updater.DownloadUpdate()
	if err != nil {
		return fmt.Sprintf("Error downloading update: %s", err.Error())
	}
	
	// Apply the update
	err = a.updater.ApplyUpdate(downloadPath, updateInfo)
	if err != nil {
		return fmt.Sprintf("Error applying update: %s", err.Error())
	}
	
	return "Update process initiated. Please follow any instructions that appear."
}
