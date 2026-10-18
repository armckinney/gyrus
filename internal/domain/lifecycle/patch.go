package lifecycle

import (
	"time"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// ApplyPatch applies the common lifecycle, schema, and revision rules to a detached concept.
func ApplyPatch(doc *gyrus.Concept, patch gyrus.ConceptPatch) error {
	if err := ValidateMutation(doc.Type, doc.Status, doc.Immutable, patch.Content != nil && *patch.Content != doc.Content); err != nil {
		return err
	}
	if patch.Status != nil {
		if err := ValidateTransition(doc.Type, doc.Status, *patch.Status); err != nil {
			return err
		}
		doc.Status = *patch.Status
	}
	if patch.Title != nil {
		doc.Title = *patch.Title
	}
	if patch.Tags != nil {
		doc.Tags = *patch.Tags
	}
	if patch.Links != nil {
		doc.Links = *patch.Links
	}
	if patch.Content != nil {
		doc.Content = *patch.Content
	}
	doc.Version++
	doc.LastUpdated = time.Now().UTC().Truncate(time.Second)
	return okf.Validate(doc)
}
