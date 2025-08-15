package gorun

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func BenchmarkRunStopProgram(b *testing.B) {
	// Build test program
	execPath := buildTestProgramBench(b, "simple_program")
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

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := gr.RunProgram()
		if err != nil {
			b.Fatalf("RunProgram() failed: %v", err)
		}

		// Let it run briefly
		time.Sleep(10 * time.Millisecond)

		err = gr.StopProgram()
		if err != nil && !strings.Contains(err.Error(), "no child processes") {
			b.Fatalf("StopProgram() failed: %v", err)
		}

		// Wait for it to stop
		time.Sleep(10 * time.Millisecond)
	}
}

func BenchmarkIsRunning(b *testing.B) {
	// Build test program
	execPath := buildTestProgramBench(b, "long_program")
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
		b.Fatalf("RunProgram() failed: %v", err)
	}
	defer gr.StopProgram()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gr.IsRunning()
	}
}

func BenchmarkNew(b *testing.B) {
	exitChan := make(chan bool)
	buf := &bytes.Buffer{}

	config := &GoRunConfig{
		ExecProgramPath: "test",
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          buf,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = New(config)
	}
}

// buildTestProgramBench builds a test program for benchmarks
func buildTestProgramBench(tb testing.TB, programName string) string {
	tb.Helper()

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
		tb.Fatalf("Failed to build test program %s: %v", programName, err)
	}

	return execPath
}
