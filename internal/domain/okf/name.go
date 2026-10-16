package okf

import (
	"strings"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// ParseName accepts only canonical collection-relative concept names.
func ParseName(name string) (tenant, project, id string, err error) {
	parts := strings.Split(name, "/")
	if len(parts) == 3 && parts[1] == "reference" {
		tenant, id = parts[0], parts[2]
	} else if len(parts) == 4 && parts[1] == "projects" {
		tenant, project, id = parts[0], parts[2], parts[3]
	} else {
		return "", "", "", &ValidationError{Field: "name", Message: "use tenant/reference/id or tenant/projects/project/id"}
	}
	if err = ValidateSegment("tenant", tenant); err != nil {
		return
	}
	if len(parts) == 4 {
		if err = ValidateSegment("project", project); err != nil {
			return
		}
	}
	err = ValidateID(id)
	return
}

func ValidateName(name string) error { _, _, _, err := ParseName(name); return err }

func ValidateEdge(edge gyrus.ConceptEdge) error {
	if err := ValidateName(edge.FromConceptID); err != nil {
		return err
	}
	if err := ValidateName(edge.ToConceptID); err != nil {
		return err
	}
	if !gyrus.ValidRelationship(edge.RelationshipType) {
		return &ValidationError{Field: "rel", Message: "invalid relationship type"}
	}
	return nil
}

// SetName derives identity from logical location without storing it in frontmatter.
func SetName(doc *gyrus.Concept, defaultProject string) error {
	relative, err := ConceptPath(*doc, defaultProject)
	if err != nil {
		return err
	}
	name := strings.TrimSuffix(strings.TrimPrefix(relative, "concepts/"), ".md")
	doc.Name = name
	_, doc.Project, _, err = ParseName(name)
	doc.Scope = "reference"
	if doc.Project != "" {
		doc.Scope = "project"
	}
	return err
}
