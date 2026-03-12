package servicemanager

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// CleanCache removes all cached service installations from the workspace.
// It skips services that are currently running and prompts for confirmation unless
// verbose mode is enabled or the user confirms the operation.
func (sm *ServiceManager) CleanCache() error {

	// Scan the installation directory
	files, err := os.ReadDir(sm.Config.TmpDir)
	if err != nil {
		return fmt.Errorf("failed to read installation directory: %s", err)
	}

	if len(files) == 0 {
		fmt.Println("No cached services found.")
		return nil
	}

	// Collect information about cached services
	type cachedService struct {
		name      string
		path      string
		version   string
		isRunning bool
		size      int64
	}

	var services []cachedService
	var totalSize int64

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		servicePath := path.Join(sm.Config.TmpDir, file.Name())

		// Try to load install file to get service info
		installFile, err := sm.Ledger.LoadInstallFile(servicePath)
		version := "unknown"
		serviceName := file.Name()

		if err == nil {
			version = installFile.Version
			serviceName = installFile.Service
		}

		// Check if service is running
		isRunning := false
		if state, err := sm.Ledger.LoadStateFile(servicePath); err == nil {
			// Get list of all running PIDs
			pids := sm.Platform.PidLookup()
			if _, exists := pids[state.Pid]; exists {
				isRunning = true
			}
		}

		// Calculate directory size
		size := calculateDirSize(servicePath)
		totalSize += size

		services = append(services, cachedService{
			name:      serviceName,
			path:      servicePath,
			version:   version,
			isRunning: isRunning,
			size:      size,
		})
	}

	if len(services) == 0 {
		fmt.Println("No cached services found.")
		return nil
	}

	// Pre-format sizes so we can measure their width
	formattedSizes := make([]string, len(services))
	for i, svc := range services {
		formattedSizes[i] = formatSize(svc.size)
	}

	// Build table and compute dynamic column widths
	tbl := NewTable("  ", "   ")
	tbl.SetMinWidth(0, 10, false)                             // name
	tbl.SetMinWidth(1, 7, false)                              // version
	tbl.SetMinWidth(2, 4, true)                               // size (right-aligned)
	tbl.SetMinWidth(3, len("[RUNNING (will skip)]")+1, false) // status
	for i, svc := range services {
		tbl.FitRow(svc.name, svc.version, formattedSizes[i])
	}

	// Display what will be pruned
	fmt.Printf("\nFound %d cached service(s):\n", len(services))
	tbl.Separator()

	runningCount := 0
	for i, svc := range services {
		status := "ready to prune"
		color := ColorGreen
		if svc.isRunning {
			status = "RUNNING (will skip)"
			color = ColorYellow
			runningCount++
		}
		tbl.PrintColoredRow(
			[]string{svc.name, svc.version, formattedSizes[i], "[" + status + "]"},
			[]string{"", "", "", color},
		)
	}

	tbl.Separator()
	fmt.Printf("Total disk space: %s\n", formatSize(totalSize))

	if runningCount > 0 {
		fmt.Printf("\nNote: %d service(s) are currently running and will be skipped.\n", runningCount)
	}

	if runningCount == len(services) {
		fmt.Println("\nAll cached services are currently running. Nothing to delete.")
		return nil
	}

	// Ask for confirmation
	fmt.Print("\nAre you sure you want to delete all cached services? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %s", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Operation cancelled.")
		return nil
	}

	// Perform the pruning
	fmt.Println("\nPruning cached services...")

	deletedCount := 0
	skippedCount := 0
	freedSpace := int64(0)

	for _, svc := range services {
		if svc.isRunning {
			sm.PrintVerbose("Skipping %s (running)\n", svc.name)
			skippedCount++
			continue
		}

		sm.PrintVerbose("Deleting %s...\n", svc.name)

		err := os.RemoveAll(svc.path)
		if err != nil {
			fmt.Printf("Warning: Failed to delete %s: %s\n", svc.name, err)
			skippedCount++
		} else {
			deletedCount++
			freedSpace += svc.size
			if !sm.Commands.NoProgress {
				fmt.Printf("  ✓ Deleted %s (%s)\n", svc.name, svc.version)
			}
		}
	}

	// Print summary
	fmt.Println("\nPrune complete:")
	fmt.Printf("  • Deleted: %d service(s)\n", deletedCount)
	fmt.Printf("  • Skipped: %d service(s)\n", skippedCount)
	fmt.Printf("  • Freed: %s\n", formatSize(freedSpace))

	if deletedCount > 0 {
		fmt.Println("\nDeleted services can be re-downloaded when needed using --start.")
	}

	return nil
}

// calculateDirSize recursively calculates the size of a directory in bytes
func calculateDirSize(dirPath string) int64 {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0
	}

	return size
}

// formatSize converts bytes to a human-readable format
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
