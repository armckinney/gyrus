package okf

import (
	"fmt"
	"regexp"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

var idRegex = regexp.MustCompile(`^[a-z0-9-_]+$`)

var validCategories = map[gyrus.Category]bool{
	gyrus.CategoryArchitecture:  true,
	gyrus.CategoryBusinessLogic: true,
	gyrus.CategoryProduct:       true,
	gyrus.CategoryOperations:    true,
	gyrus.CategoryTechnical:     true,
}

var validTypes = map[gyrus.DocumentType]bool{
	gyrus.TypeADR:                 true,
	gyrus.TypePRD:                 true,
	gyrus.TypeGuide:               true,
	gyrus.TypeImprovementProposal: true,
	gyrus.TypeReleaseNote:         true,
	gyrus.TypeSpecification:       true,
	gyrus.TypeStandards:           true,
	gyrus.TypeTechnicalReference:  true,
	gyrus.TypeProduct:             true,
	gyrus.TypeGlossary:            true,
	gyrus.TypeFreeform:            true,
}

// ValidationError represents an OKF schema rule violation.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error [%s]: %s", e.Field, e.Message)
}

// Validate checks a Document against mandatory OKF schema boundaries.
func Validate(doc *gyrus.Document) error {
	if doc == nil {
		return &ValidationError{Field: "document", Message: "document cannot be nil"}
	}
	if doc.ID == "" {
		return &ValidationError{Field: "id", Message: "id is mandatory"}
	}
	if !idRegex.MatchString(doc.ID) {
		return &ValidationError{Field: "id", Message: fmt.Sprintf("id '%s' must match pattern ^[a-z0-9-_]+$", doc.ID)}
	}
	if doc.Title == "" {
		return &ValidationError{Field: "title", Message: "title is mandatory and cannot be empty"}
	}
	if doc.OwnerGroup == "" {
		return &ValidationError{Field: "owner_group", Message: "owner_group is mandatory and cannot be empty"}
	}
	if !validCategories[doc.Category] {
		return &ValidationError{Field: "category", Message: fmt.Sprintf("invalid category '%s'", doc.Category)}
	}
	if !validTypes[doc.Type] {
		return &ValidationError{Field: "type", Message: fmt.Sprintf("invalid document type '%s'", doc.Type)}
	}
	if doc.Version < 1 {
		return &ValidationError{Field: "version", Message: "version must be >= 1"}
	}
	if doc.Status == "" {
		return &ValidationError{Field: "status", Message: "status is mandatory and cannot be empty"}
	}
	return nil
}
