package gorun

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestStopProgram_Basic(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &GoRunConfig{
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

	if gr.IsRunning() {
		t.Error("Program should not be running after StopProgram()")
	}
}

func TestStopProgram_WhenNotRunning(t *testing.T) {
	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &GoRunConfig{
		ExecProgramPath: "test",
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

	if gr.IsRunning() {
		t.Error("Program should not be running")
	}
}

func TestStopProgram_GracefulShutdown(t *testing.T) {
	// Build test program that handles signals
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &GoRunConfig{
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

	// Wait for it to start
	time.Sleep(50 * time.Millisecond)

	// Stop the program
	err = gr.StopProgram()
	if err != nil && !strings.Contains(err.Error(), "no child processes") {
		t.Errorf("StopProgram() failed with unexpected error: %v", err)
	}

	// Wait for it to process the signal
	time.Sleep(200 * time.Millisecond)

	output := gr.GetOutput()

	// Check that the program completed (either gracefully or was terminated)
	// The specific signal handling may vary by OS and implementation
	if !strings.Contains(output, "SIGNAL_RECEIVED") && !strings.Contains(output, "closed") {
		t.Errorf("Expected either signal handling or program completion. Output: %s", output)
	}

	if gr.IsRunning() {
		t.Error("Program should not be running after shutdown")
	}
}

func TestStopProgram_ForcedTermination(t *testing.T) {
	// Build a program that doesn't handle signals (should be force-killed)
	execPath := buildTestProgram(t, "long_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &GoRunConfig{
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

	// Wait for it to start
	time.Sleep(50 * time.Millisecond)

	if !gr.IsRunning() {
		t.Error("Program should be running")
	}

	// Stop the program - ignore "no child processes" error which can happen
	// if the process exits on its own before we stop it
	err = gr.StopProgram()
	if err != nil && !strings.Contains(err.Error(), "no child processes") {
		t.Errorf("StopProgram() failed with unexpected error: %v", err)
	}

	// Wait for the termination process to complete
	// The implementation should try graceful first, then force kill
	time.Sleep(300 * time.Millisecond)

	if gr.IsRunning() {
		t.Error("Program should not be running after forced termination")
	}
}

func TestStopProgram_MultipleCalls(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &GoRunConfig{
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

	// Call StopProgram multiple times
	err1 := gr.StopProgram()
	err2 := gr.StopProgram()
	err3 := gr.StopProgram()

	if err1 != nil {
		t.Errorf("First StopProgram() failed: %v", err1)
	}

	// Subsequent calls should not fail
	if err2 != nil {
		t.Errorf("Second StopProgram() failed: %v", err2)
	}

	if err3 != nil {
		t.Errorf("Third StopProgram() failed: %v", err3)
	}

	// Wait for termination
	time.Sleep(100 * time.Millisecond)

	if gr.IsRunning() {
		t.Error("Program should not be running after multiple StopProgram calls")
	}
}

func TestStopProgram_ConcurrentCalls(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "long_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &GoRunConfig{
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

	// Wait for it to start
	time.Sleep(50 * time.Millisecond)

	// Call StopProgram concurrently from multiple goroutines
	done := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func() {
			done <- gr.StopProgram()
		}()
	}

	// Collect results
	var unexpectedErrors []error
	for i := 0; i < 3; i++ {
		if err := <-done; err != nil {
			// Ignore "no child processes" errors as they can happen when
			// multiple calls try to stop an already stopped process
			if !strings.Contains(err.Error(), "no child processes") {
				unexpectedErrors = append(unexpectedErrors, err)
			}
		}
	}

	// Should not have unexpected errors (concurrent calls should be handled safely)
	if len(unexpectedErrors) > 0 {
		t.Errorf("Concurrent StopProgram calls produced unexpected errors: %v", unexpectedErrors)
	}

	// Wait for termination
	time.Sleep(200 * time.Millisecond)

	if gr.IsRunning() {
		t.Error("Program should not be running after concurrent StopProgram calls")
	}
}
