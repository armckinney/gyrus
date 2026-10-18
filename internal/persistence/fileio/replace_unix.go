//go:build !windows

package fileio

import "os"

func replace(source, target string) error { return os.Rename(source, target) }
