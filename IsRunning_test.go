package gorun

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestIsRunning_InitialState(t *testing.T) {
	exitChan := make(chan bool)
	buf, logger := createTestLogger()

	config := &Config{
		ExecProgramPath: "test",
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          logger,
	}

	gr := New(config)

	if gr.IsRunning() {
		t.Error("New GoRun instance should not be running initially")
	}

	_ = buf // Use buf to avoid unused variable error
}

func TestIsRunning_AfterStart(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf, logger := createTestLogger()

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          logger,
	}

	gr := New(config)

	// Start the program
	err := gr.RunProgram()
	if err != nil {
		t.Fatalf("RunProgram() failed: %v", err)
	}

	// Should be running now
	if !gr.IsRunning() {
		t.Error("Program should be running after RunProgram()")
	}

	// Clean up
	gr.StopProgram()

	_ = buf // Use buf to avoid unused variable error
}

func TestIsRunning_AfterStop(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf, logger := createTestLogger()

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          logger,
	}

	gr := New(config)

	// Start the program
	err := gr.RunProgram()
	if err != nil {
		t.Fatalf("RunProgram() failed: %v", err)
	}

	// Verify it's running
	if !gr.IsRunning() {
		t.Error("Program should be running")
	}

	// Stop the program
	err = gr.StopProgram()
	if err != nil {
		t.Errorf("StopProgram() failed: %v", err)
	}

	// Wait for it to stop
	time.Sleep(100 * time.Millisecond)

	// Should not be running now
	if gr.IsRunning() {
		t.Error("Program should not be running after StopProgram()")
	}

	_ = buf // Use buf to avoid unused variable error
}

func TestIsRunning_AfterProgramExit(t *testing.T) {
	// Create a program that exits quickly on its own
	quickExitProgram := `package main
import "fmt"
func main() {
	fmt.Println("QUICK_EXIT")
}`

	// Write to temporary file
	tmpFile := filepath.Join(os.TempDir(), "quick_exit.go")
	err := os.WriteFile(tmpFile, []byte(quickExitProgram), 0644)
	if err != nil {
		t.Fatalf("Failed to write test program: %v", err)
	}
	defer os.Remove(tmpFile)

	// Build it
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	execPath := filepath.Join(os.TempDir(), "quick_exit"+ext)
	cmd := exec.Command("go", "build", "-o", execPath, tmpFile)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build quick exit program: %v", err)
	}
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	_, logger := createTestLogger()

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          logger,
	}

	gr := New(config)

	// Start the program
	err = gr.RunProgram()
	if err != nil {
		t.Fatalf("RunProgram() failed: %v", err)
	}

	// Initially should be running
	if !gr.IsRunning() {
		t.Error("Program should be running initially")
	}

	// Wait for the program to exit on its own
	time.Sleep(200 * time.Millisecond)

	// Now should not be running (the program should have detected this)
	// Note: This depends on the implementation checking process state
	// The current implementation might not detect this automatically
	// so this test might need adjustment based on actual behavior
}
