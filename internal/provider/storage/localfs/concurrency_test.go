package localfs_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestLocalConcurrentWriters(t *testing.T) {
	if root := os.Getenv("GYRUS_LOCK_CHILD"); root != "" {
		store, err := localfs.NewStore(root)
		if err != nil {
			os.Exit(8)
		}
		title := fmt.Sprintf("writer-%d", os.Getpid())
		_, err = store.Update(context.Background(), "root/reference/shared", gyrus.ConceptPatch{Title: &title}, 1)
		if err == nil {
			os.Exit(0)
		}
		var conflict *gyrus.ConflictError
		if errors.As(err, &conflict) {
			os.Exit(4)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(8)
	}
	root := t.TempDir()
	store, _ := localfs.NewStore(root)
	_, err := store.Create(context.Background(), gyrus.Concept{ID: "shared", Title: "original", Tenant: "root", Category: gyrus.CategoryTechnical, Type: gyrus.TypeGuide, Version: 1, Status: "active", Name: "root/reference/shared"})
	if err != nil {
		t.Fatal(err)
	}
	commands := make([]*exec.Cmd, 2)
	for i := range commands {
		commands[i] = exec.Command(os.Args[0], "-test.run=^TestLocalConcurrentWriters$")
		commands[i].Env = append(os.Environ(), "GYRUS_LOCK_CHILD="+root)
		if err := commands[i].Start(); err != nil {
			t.Fatal(err)
		}
	}
	wins, conflicts := 0, 0
	for _, cmd := range commands {
		err := cmd.Wait()
		if err == nil {
			wins++
		} else if cmd.ProcessState.ExitCode() == 4 {
			conflicts++
		} else {
			t.Fatalf("unexpected writer error: %v", err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("wins=%d conflicts=%d", wins, conflicts)
	}
	doc, err := store.Get(context.Background(), "root/reference/shared")
	if err != nil || doc.Version != 2 {
		t.Fatalf("winner payload: %+v %v", doc, err)
	}
}
