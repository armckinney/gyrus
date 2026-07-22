package sqlite

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/armckinney/gyrus/internal/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// Sync performs incremental file re-indexing over storageRoot.
func (idx *Indexer) Sync(ctx context.Context, storageRoot string) (gyrus.SyncReport, error) {
	report := gyrus.SyncReport{}

	absRoot, err := filepath.Abs(storageRoot)
	if err != nil {
		return report, fmt.Errorf("invalid sync root path: %w", err)
	}

	scannedOnDisk := make(map[string]string) // id -> hash

	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			report.ScannedFiles++

			data, err := os.ReadFile(path)
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("failed reading %s: %v", path, err))
				return nil
			}

			doc, err := okf.ParseMarkdown(data)
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("invalid OKF markdown in %s: %v", path, err))
				return nil
			}

			hashBytes := sha256.Sum256(data)
			checksum := hex.EncodeToString(hashBytes[:])
			scannedOnDisk[doc.ID] = checksum

			// Check existing indexed checksum
			var existingChecksum string
			err = idx.db.QueryRowContext(ctx, `SELECT checksum FROM documents_index WHERE id = ?`, doc.ID).Scan(&existingChecksum)

			if err != nil || existingChecksum != checksum {
				if err := idx.Index(ctx, *doc); err != nil {
					report.Errors = append(report.Errors, fmt.Sprintf("indexing failed for %s: %v", doc.ID, err))
				} else {
					report.IndexedFiles++
					// Update checksum record
					_, _ = idx.db.ExecContext(ctx, `UPDATE documents_index SET checksum = ? WHERE id = ?`, checksum, doc.ID)

					// Auto-extract relationship edges from dependencies
					if len(doc.Dependencies) > 0 {
						var edges []gyrus.DocumentEdge
						for _, depID := range doc.Dependencies {
							edges = append(edges, gyrus.DocumentEdge{
								FromDocumentID:   doc.ID,
								ToDocumentID:     depID,
								RelationshipType: gyrus.RelDependsOn,
								CreatedBy:        doc.OwnerGroup,
							})
						}
						_ = idx.UpsertEdges(ctx, edges)
					}
				}
			} else {
				report.UnchangedFiles++
			}
		}
		return nil
	})

	if err != nil {
		return report, fmt.Errorf("sync walk failed: %w", err)
	}

	// Purge stale index records for files deleted on disk
	rows, err := idx.db.QueryContext(ctx, `SELECT id FROM documents_index`)
	if err == nil {
		defer rows.Close()
		var dbIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err == nil {
				dbIDs = append(dbIDs, id)
			}
		}
		for _, dbID := range dbIDs {
			if _, exists := scannedOnDisk[dbID]; !exists {
				_ = idx.Remove(ctx, dbID)
				report.RemovedFiles++
			}
		}
	}

	return report, nil
}
