package fileio

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestLockCancellationAndProcessExit(t *testing.T) {
	if root := os.Getenv("GYRUS_HOLD_LOCK"); root != "" {
		unlock, err := Lock(context.Background(), root)
		if err != nil {
			os.Exit(8)
		}
		defer unlock()
		fmt.Println("locked")
		time.Sleep(time.Minute)
		os.Exit(0)
	}
	root := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestLockCancellationAndProcessExit$")
	cmd.Env = append(os.Environ(), "GYRUS_HOLD_LOCK="+root)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	if line, err := bufio.NewReader(stdout).ReadString('\n'); err != nil || line != "locked\n" {
		t.Fatalf("child ready: %q %v", line, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer cancel()
	if _, err := Lock(ctx, root); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock cancellation: %v", err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait()
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	unlock, err := Lock(ctx2, root)
	if err != nil {
		t.Fatal("lock not released on process exit", err)
	}
	unlock()
}
