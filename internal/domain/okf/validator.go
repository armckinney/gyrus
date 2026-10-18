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

var validTypes = map[gyrus.ConceptType]bool{
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

// Validate checks a Concept against mandatory OKF schema boundaries.
func Validate(concept *gyrus.Concept) error {
	if concept == nil {
		return &ValidationError{Field: "concept", Message: "concept cannot be nil"}
	}
	if concept.Name != "" {
		if err := ValidateName(concept.Name); err != nil {
			return err
		}
	}
	if concept.ID == "" {
		return &ValidationError{Field: "id", Message: "id is mandatory"}
	}
	if !idRegex.MatchString(concept.ID) {
		return &ValidationError{Field: "id", Message: fmt.Sprintf("id '%s' must match pattern ^[a-z0-9-_]+$", concept.ID)}
	}
	if err := ValidateSegment("id", concept.ID); err != nil {
		return err
	}
	if concept.Title == "" {
		return &ValidationError{Field: "title", Message: "title is mandatory and cannot be empty"}
	}
	if concept.Tenant == "" {
		return &ValidationError{Field: "tenant", Message: "tenant is mandatory and cannot be empty"}
	}
	if !validCategories[concept.Category] {
		return &ValidationError{Field: "category", Message: fmt.Sprintf("invalid category '%s'", concept.Category)}
	}
	if !validTypes[concept.Type] {
		return &ValidationError{Field: "type", Message: fmt.Sprintf("invalid concept type '%s'", concept.Type)}
	}
	if concept.Version < 1 {
		return &ValidationError{Field: "version", Message: "version must be >= 1"}
	}
	if concept.Status == "" {
		return &ValidationError{Field: "status", Message: "status is mandatory and cannot be empty"}
	}
	for _, link := range concept.Links {
		if err := ValidateName(link.Concept); err != nil {
			return &ValidationError{Field: "links.concept", Message: err.Error()}
		}
		if !gyrus.ValidRelationship(link.Rel) {
			return &ValidationError{Field: "links.rel", Message: "invalid relationship type"}
		}
	}
	if err := ValidateSegment("tenant", concept.Tenant); err != nil {
		return err
	}
	if concept.Project != "" {
		if err := ValidateSegment("project", concept.Project); err != nil {
			return err
		}
	}
	return nil
}
