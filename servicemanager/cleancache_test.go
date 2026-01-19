package servicemanager

import (
	"os"
	"path"
	"testing"

	"sm2/cli"
	"sm2/ledger"
	"sm2/platform"
)

// TestCleanCache_EmptyDirectory tests pruning when no services are cached
func TestCleanCache_EmptyDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	sm := ServiceManager{
		Config: ServiceManagerConfig{
			TmpDir: tmpDir,
		},
		Commands: cli.UserOption{
			Verbose: true, // Skip confirmation prompt
		},
		Ledger: ledger.NewLedger(),
	}

	err := sm.CleanCache()
	if err != nil {
		t.Errorf("Expected no error for empty directory, got: %s", err)
	}
}

// TestCleanCache_WithCachedServices tests pruning with cached services
func TestCleanCache_WithCachedServices(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a fake service directory
	serviceDir := path.Join(tmpDir, "TEST_SERVICE")
	err := os.MkdirAll(serviceDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test service directory: %s", err)
	}

	// Create an install file for the service
	installFile := ledger.InstallFile{
		Service: "TEST_SERVICE",
		Version: "1.0.0",
		Path:    serviceDir,
	}

	testLedger := ledger.NewLedger()
	err = testLedger.SaveInstallFile(serviceDir, installFile)
	if err != nil {
		t.Fatalf("Failed to save install file: %s", err)
	}

	sm := ServiceManager{
		Config: ServiceManagerConfig{
			TmpDir: tmpDir,
		},
		Commands: cli.UserOption{
			Verbose: true, // Skip confirmation prompt
		},
		Ledger: testLedger,
		Platform: platform.Platform{
			PidLookup: func() map[int]int {
				// Return map with current PID (simulating running process)
				return map[int]int{os.Getpid(): os.Getpid()}
			},
		},
	}

	err = sm.CleanCache()
	if err != nil {
		t.Errorf("Expected no error, got: %s", err)
	}

	// Verify the service directory was deleted
	if _, err := os.Stat(serviceDir); !os.IsNotExist(err) {
		t.Errorf("Expected service directory to be deleted")
	}
}

// TestCleanCache_SkipsRunningServices tests that running services are not deleted
func TestCleanCache_SkipsRunningServices(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a fake service directory
	serviceDir := path.Join(tmpDir, "RUNNING_SERVICE")
	err := os.MkdirAll(serviceDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test service directory: %s", err)
	}

	// Create an install file
	installFile := ledger.InstallFile{
		Service: "RUNNING_SERVICE",
		Version: "1.0.0",
		Path:    serviceDir,
	}

	testLedger := ledger.NewLedger()
	err = testLedger.SaveInstallFile(serviceDir, installFile)
	if err != nil {
		t.Fatalf("Failed to save install file: %s", err)
	}

	// Create a state file with the current process PID (this process is running)
	stateFile := ledger.StateFile{
		Service: "RUNNING_SERVICE",
		Version: "1.0.0",
		Pid:     os.Getpid(), // Use current process PID
		Path:    serviceDir,
	}

	err = testLedger.SaveStateFile(serviceDir, stateFile)
	if err != nil {
		t.Fatalf("Failed to save state file: %s", err)
	}

	sm := ServiceManager{
		Config: ServiceManagerConfig{
			TmpDir: tmpDir,
		},
		Commands: cli.UserOption{
			Verbose: true,
		},
		Ledger: testLedger,
		Platform: platform.Platform{
			PidLookup: func() map[int]int {
				// Return map with current PID (simulating running process)
				return map[int]int{os.Getpid(): os.Getpid()}
			},
		},
	}

	err = sm.CleanCache()
	if err != nil {
		t.Errorf("Expected no error, got: %s", err)
	}

	// Verify the service directory was NOT deleted (because it's "running")
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		t.Errorf("Expected running service directory to be preserved")
	}
}

// TestCleanCache_InvalidPath tests error handling for invalid paths
func TestCleanCache_InvalidPath(t *testing.T) {
	sm := ServiceManager{
		Config: ServiceManagerConfig{
			TmpDir: "relative/path", // Relative path should fail
		},
		Commands: cli.UserOption{
			Verbose: true,
		},
		Ledger: ledger.NewLedger(),
	}

	err := sm.CleanCache()
	if err == nil {
		t.Error("Expected error for relative path, got nil")
	}
}

// TestFormatSize tests the human-readable size formatting
func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 bytes"},
		{500, "500 bytes"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1572864, "1.50 MB"},
		{1073741824, "1.00 GB"},
		{1610612736, "1.50 GB"},
	}

	for _, test := range tests {
		result := formatSize(test.bytes)
		if result != test.expected {
			t.Errorf("formatSize(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

// TestCalculateDirSize tests directory size calculation
func TestCalculateDirSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	testFile1 := path.Join(tmpDir, "file1.txt")
	testFile2 := path.Join(tmpDir, "file2.txt")

	content1 := []byte("Hello, World!")
	content2 := []byte("Test content")

	err := os.WriteFile(testFile1, content1, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %s", err)
	}

	err = os.WriteFile(testFile2, content2, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %s", err)
	}

	size := calculateDirSize(tmpDir)
	expectedSize := int64(len(content1) + len(content2))

	if size != expectedSize {
		t.Errorf("calculateDirSize() = %d, expected %d", size, expectedSize)
	}
}
