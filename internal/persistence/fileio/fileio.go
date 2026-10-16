// Package fileio coordinates cooperating local writers and replaces complete payloads.
package fileio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

const LockRetryInterval = 25 * time.Millisecond
const FileMode = 0644
const DirectoryMode = 0755

// Lock acquires a root-wide interprocess lock until the returned release is called.
func Lock(ctx context.Context, root string) (func(), error) {
	lock := flock.New(filepath.Join(root, ".gyrus-write.lock"))
	ok, err := lock.TryLockContext(ctx, LockRetryInterval)
	if err != nil {
		_ = lock.Close()
		return nil, err
	}
	if !ok {
		_ = lock.Close()
		return nil, ctx.Err()
	}
	return func() { _ = lock.Unlock(); _ = lock.Close() }, nil
}

// Write replaces a file only after its complete new payload has been flushed.
func Write(name string, data []byte) error {
	mode := os.FileMode(FileMode)
	if info, err := os.Stat(name); err == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(name), DirectoryMode); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(name), ".gyrus-tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(mode); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return replace(f.Name(), name)
}

// SafePath rejects paths escaping root and existing symlink components.
func SafePath(root, key string) (string, error) {
	if !filepath.IsLocal(filepath.FromSlash(key)) {
		return "", fmt.Errorf("path escapes storage root")
	}
	name := root
	for _, component := range strings.Split(filepath.FromSlash(key), string(filepath.Separator)) {
		name = filepath.Join(name, component)
		info, err := os.Lstat(name)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("storage path contains symlink: %s", name)
		}
	}
	return name, nil
}
