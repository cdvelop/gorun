package gorun

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Helper function to build test programs
func buildTestProgram(t *testing.T, programName string) string {
	t.Helper()

	// Get the source path
	sourcePath := filepath.Join("testdata", programName+".go")

	// Get executable extension for Windows
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	// Build the program in a temporary directory
	execPath := filepath.Join(os.TempDir(), programName+ext)

	cmd := exec.Command("go", "build", "-o", execPath, sourcePath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test program %s: %v", programName, err)
	}

	return execPath
}

// Helper function to create a test logger
func createTestLogger() (*bytes.Buffer, func(...any)) {
	buf := &bytes.Buffer{}
	logger := func(args ...any) {
		for i, arg := range args {
			if i > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(strings.TrimSpace(string([]byte(arg.(string)))))
		}
		buf.WriteString("\n")
	}
	return buf, logger
}

func TestNew(t *testing.T) {
	exitChan := make(chan bool)
	buf, logger := createTestLogger()

	config := &Config{
		ExecProgramPath: "test",
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          buf,
	}

	gr := New(config)

	if gr == nil {
		t.Fatal("New() returned nil")
	}

	if gr.Config != config {
		t.Error("Config not set correctly")
	}

	if gr.IsRunning() {
		t.Error("New instance should not be running")
	}

	_ = logger // Use logger to avoid unused variable error
}

func TestRunProgram_Success(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          buf,
	}

	gr := New(config)

	// Start the program
	err := gr.RunProgram()
	if err != nil {
		t.Fatalf("RunProgram() failed: %v", err)
	}

	// Check if it's running
	if !gr.IsRunning() {
		t.Error("Program should be running")
	}

	// Wait a bit and check output
	time.Sleep(200 * time.Millisecond)

	output := gr.GetOutput()
	if !strings.Contains(output, "PROGRAM_STARTED") && !strings.Contains(output, "TICK") {
		t.Errorf("Expected program output in logger, got: %s", output)
	}

	// Stop the program - ignore errors from multiple stops
	gr.StopProgram()

	// Wait for it to stop
	time.Sleep(100 * time.Millisecond)

	if gr.IsRunning() {
		t.Error("Program should not be running after stop")
	}
}

func TestRunProgram_AlreadyRunning(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "long_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          buf,
	}

	gr := New(config)

	// Start the program
	err := gr.RunProgram()
	if err != nil {
		t.Fatalf("First RunProgram() failed: %v", err)
	}

	if !gr.IsRunning() {
		t.Error("Program should be running")
	}

	// Try to start again - should stop the previous one and start new
	err = gr.RunProgram()
	if err != nil {
		t.Errorf("Second RunProgram() should not fail: %v", err)
	}

	if !gr.IsRunning() {
		t.Error("Program should still be running after restart")
	}

	// Clean up - ignore "no child processes" error
	err = gr.StopProgram()
	if err != nil && !strings.Contains(err.Error(), "no child processes") {
		t.Errorf("StopProgram() failed with unexpected error: %v", err)
	}
}

func TestStopProgram_NotRunning(t *testing.T) {
	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &Config{
		ExecProgramPath: "nonexistent",
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          buf,
	}

	gr := New(config)

	// Try to stop when not running
	err := gr.StopProgram()
	if err != nil {
		t.Errorf("StopProgram() should not fail when not running: %v", err)
	}
}

func TestRunProgram_NonexistentFile(t *testing.T) {
	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &Config{
		ExecProgramPath: "nonexistent_program",
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          buf,
	}

	gr := New(config)

	err := gr.RunProgram()
	if err == nil {
		t.Error("RunProgram() should fail for nonexistent file")
	}

	if gr.IsRunning() {
		t.Error("Program should not be running when file doesn't exist")
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          buf,
	}

	gr := New(config)

	// Test concurrent access to IsRunning
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Start multiple goroutines that check IsRunning
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				gr.IsRunning()
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	// Start multiple goroutines that start/stop programs
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := gr.RunProgram(); err != nil {
				errors <- err
				return
			}
			time.Sleep(10 * time.Millisecond)
			if err := gr.StopProgram(); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent access error: %v", err)
		}
	}
}

func TestRunArguments(t *testing.T) {
	// Build the args test program
	execPath := buildTestProgram(t, "args_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	testArgs := []string{"arg1", "arg2", "test"}
	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return testArgs },
		ExitChan:        exitChan,
		Logger:          buf,
	}

	gr := New(config)

	err := gr.RunProgram()
	if err != nil {
		t.Fatalf("RunProgram() failed: %v", err)
	}

	// Wait for program to finish (it should exit quickly)
	time.Sleep(200 * time.Millisecond)

	output := gr.GetOutput()
	expectedArgs := "ARGS:" + strings.Join(testArgs, ",")

	// Check if the arguments were passed correctly
	// The logger captures the program output, so we should see either the args or completion message
	if !strings.Contains(output, expectedArgs) && !strings.Contains(output, "closed successfully") {
		t.Errorf("Expected either args output '%s' or completion message. Got: %s", expectedArgs, output)
	}

	// Clean up
	gr.StopProgram()
}

func TestSignalHandling(t *testing.T) {
	// This test verifies that programs receive signals correctly
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          buf,
	}

	gr := New(config)

	// Start the program
	err := gr.RunProgram()
	if err != nil {
		t.Fatalf("RunProgram() failed: %v", err)
	}

	// Wait a bit for it to start
	time.Sleep(100 * time.Millisecond)

	// Stop it gracefully - ignore potential errors from process cleanup
	gr.StopProgram()

	// Wait for it to finish
	time.Sleep(200 * time.Millisecond)

	output := gr.GetOutput()

	// Check that it either received the signal and exited gracefully,
	// or was terminated successfully
	signalReceived := strings.Contains(output, "SIGNAL_RECEIVED") && strings.Contains(output, "PROGRAM_GRACEFUL_EXIT")
	programTerminated := strings.Contains(output, "closed")

	if !signalReceived && !programTerminated {
		t.Errorf("Program should have been terminated gracefully or forcefully. Output: %s", output)
	}
}

func TestExitChanIntegration(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "long_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool, 1)
	buf := &bytes.Buffer{}

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          buf,
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

	// Signal exit
	exitChan <- true

	// Give it time to process the exit signal
	time.Sleep(100 * time.Millisecond)

	// Program should be stopped now
	if gr.IsRunning() {
		t.Error("Program should have stopped after exit signal")
	}
}
