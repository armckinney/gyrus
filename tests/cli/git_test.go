package cli_test

import (
	"context"
	"fmt"
	storagegit "github.com/armckinney/gyrus/internal/provider/storage/git"
	"github.com/armckinney/gyrus/pkg/gyrus"
	gogit "github.com/go-git/go-git/v5"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIGitRuntimeBranch(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "remote.git")
	if _, err := gogit.PlainInit(remote, true); err != nil {
		t.Fatal(err)
	}
	store, err := storagegit.NewStore(storagegit.Options{RepoURL: remote})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Create(context.Background(), gyrus.Concept{ID: "shared", Title: "shared", Type: gyrus.TypeGuide, Category: gyrus.CategoryTechnical, Tenant: "root", Status: "active"}); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	name := filepath.Join(root, ".gyrus.yaml")
	original := fmt.Sprintf("storage:\n  provider: git\ngit:\n  repo_url: %q\n  commit_strategy: branching\n", remote)
	if err = os.WriteFile(name, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(wantError bool, args ...string) string {
		t.Helper()
		cmd := exec.Command(gyrusBinPath, args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if (err != nil) != wantError {
			t.Fatalf("%v: %v %s", args, err, out)
		}
		return string(out)
	}
	if !strings.Contains(run(false, "git", "status", "--json"), `"selection_required":true`) {
		t.Fatal("missing selection status")
	}
	run(true, "get", "root/reference/shared")
	run(false, "get", "root/reference/shared", "--git-branch", "initiative/a")
	run(false, "get", "root/reference/shared", "--git-branch", "initiative/b")
	run(true, "get", "root/reference/shared", "--git-branch", "main")
	run(true, "get", "root/reference/shared", "--git-branch", "")
	data, err := os.ReadFile(name)
	if err != nil || string(data) != original {
		t.Fatal("runtime selection modified configuration")
	}
}
