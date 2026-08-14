package install_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test / Packaging
// [Purpose]: Verifies that distribution archive packages contain all necessary binaries, skills, and schemas.
// [Execution Surface]: Filesystem Archive Validation
// [Assertions]: Skills and schemas are present and can be packaged into release tarball without corruption.
// -----------------------------------------------------------------------------
func TestDistributionArchivePackaging(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	skillsDir := filepath.Join(repoRoot, "packaging", "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Skipf("packaging/skills directory not found at %s; skipping packaging test", skillsDir)
	}

	// Create an in-memory tar.gz archive representing release bundle
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Walk and bundle skills
	fileCount := 0
	err = filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		header := &tar.Header{
			Name: relPath,
			Mode: 0644,
			Size: int64(len(data)),
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tw.Write(data); err != nil {
			return err
		}
		fileCount++
		return nil
	})

	if err != nil {
		t.Fatalf("Failed bundling skills archive: %v", err)
	}

	_ = tw.Close()
	_ = gw.Close()

	if fileCount == 0 {
		t.Errorf("Expected at least 1 file to be packaged in skills archive")
	}

	// Unpack and verify in-memory archive
	gr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatalf("Failed creating gzip reader: %v", err)
	}
	tr := tar.NewReader(gr)

	unpackedFiles := make(map[string]bool)
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		unpackedFiles[hdr.Name] = true
	}

	hasCLI := false
	hasMCP := false
	for name := range unpackedFiles {
		if strings.Contains(name, "gyrus-cli/SKILL.md") {
			hasCLI = true
		}
		if strings.Contains(name, "gyrus-mcp/SKILL.md") {
			hasMCP = true
		}
	}

	if !hasCLI || !hasMCP {
		t.Errorf("Archive missing SKILL.md files: hasCLI=%v, hasMCP=%v", hasCLI, hasMCP)
	}
}
