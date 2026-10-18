package cli_test

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestProcessRejectsMalformedUTF8FromStdin(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}

	executableName := "youngkit"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	executablePath := filepath.Join(t.TempDir(), executableName)

	buildContext, cancelBuild := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelBuild()

	build := exec.CommandContext(buildContext, "go", "build", "-o", executablePath, ".")
	build.Dir = root
	buildOutput, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, buildOutput)
	}

	processContext, cancelProcess := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelProcess()

	program := exec.CommandContext(processContext, executablePath, "rsk", "-")
	program.Stdin = bytes.NewReader([]byte{0x31, 0x20, 0xff, 0x0a})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	program.Stdout = &stdout
	program.Stderr = &stderr

	err = program.Run()
	if err == nil {
		t.Fatal("program succeeded, want exit code 2")
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("program error = %v, want an exit error", err)
	}
	if got := exitError.ExitCode(); got != 2 {
		t.Fatalf("program exit code = %d, want 2; stderr=%q", got, stderr.String())
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	const wantStderr = "youngkit: <stdin>: input is not valid UTF-8\n"
	if got := stderr.String(); got != wantStderr {
		t.Fatalf("stderr = %q, want %q", got, wantStderr)
	}
}
