package gorun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWorkingDir(t *testing.T) {
	// Create a temporary test structure
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Error creating subdirectory: %v", err)
	}

	exitChan := make(chan bool, 1)

	// Use 'pwd' command to test working directory
	config := &Config{
		ExecProgramPath: "pwd",
		ExitChan:        exitChan,
		Logger:          nil,    // use internal SafeBuffer and read via GetOutput()
		WorkingDir:      subDir, // Set working directory to subdirectory
	}

	gorun := New(config)

	t.Run("WorkingDir is set correctly", func(t *testing.T) {
		err := gorun.RunProgram()
		if err != nil {
			t.Fatalf("Error running program: %v", err)
		}

		// Wait a bit for the program to execute and output
		time.Sleep(100 * time.Millisecond)

		// Stop the program
		gorun.StopProgram()

		// Check if the output contains the expected working directory (thread-safe)
		outputStr := gorun.getOutput()
		if !strings.Contains(outputStr, subDir) {
			t.Errorf("Expected working directory %s in output, got: %s", subDir, outputStr)
		}
	})
}

func TestWorkingDirNotSet(t *testing.T) {
	// Test that when WorkingDir is not set, it uses the default working directory
	exitChan := make(chan bool, 1)

	config := &Config{
		ExecProgramPath: "pwd",
		ExitChan:        exitChan,
		Logger:          nil,
		// WorkingDir not set - should use current directory
	}

	gorun := New(config)

	t.Run("WorkingDir not set uses current directory", func(t *testing.T) {
		err := gorun.RunProgram()
		if err != nil {
			t.Fatalf("Error running program: %v", err)
		}

		// Wait a bit for the program to execute and output
		time.Sleep(100 * time.Millisecond)

		// Stop the program
		gorun.StopProgram()

		// Get current working directory
		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Error getting current directory: %v", err)
		}

		// Check if the output contains the current working directory (thread-safe)
		outputStr := gorun.getOutput()
		if !strings.Contains(outputStr, currentDir) {
			t.Errorf("Expected current directory %s in output, got: %s", currentDir, outputStr)
		}
	})
}
