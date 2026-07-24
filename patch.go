package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("/workspaces/gyrus/internal/provider/git/store.go")
	s := string(b)
	s = strings.Replace(s, "st = memory.NewStorage(); fs = memfs.New(); repo, err = git.Init(st, fs)", "st = memory.NewStorage(); fs = memfs.New(); store.fs = fs; repo, err = git.Init(st, fs)", 1)
	os.WriteFile("/workspaces/gyrus/internal/provider/git/store.go", []byte(s), 0644)
}
