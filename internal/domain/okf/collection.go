package okf

import (
	"context"
	"fmt"
	"path"
	"strings"
	"unicode"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// ValidateID validates the identifier used by ID-only public APIs.
func ValidateID(id string) error {
	if !idRegex.MatchString(id) {
		return &ValidationError{Field: "id", Message: "must match ^[a-z0-9-_]+$"}
	}
	return ValidateSegment("id", id)
}

// ValidateSegment rejects path components unsafe on any supported platform.
func ValidateSegment(name, value string) error {
	if value == "" || value == "." || value == ".." || strings.ContainsAny(value, "/\\:<>\"|?*") || strings.HasSuffix(value, ".") || strings.HasSuffix(value, " ") {
		return &ValidationError{Field: name, Message: "must be a nonempty safe path segment"}
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return &ValidationError{Field: name, Message: "must not contain control characters"}
		}
	}
	base := strings.ToUpper(strings.SplitN(value, ".", 2)[0])
	switch base {
	case "CON", "PRN", "AUX", "NUL", "CONIN$", "CONOUT$":
		return &ValidationError{Field: name, Message: "must not be a reserved device name"}
	}
	if len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9' {
		return &ValidationError{Field: name, Message: "must not be a reserved device name"}
	}
	return nil
}

// ConceptPath builds the relative canonical concept path.
func ConceptPath(doc gyrus.Concept, defaultProject string) (string, error) {
	if err := ValidateID(doc.ID); err != nil {
		return "", err
	}
	if err := ValidateSegment("tenant", doc.Tenant); err != nil {
		return "", err
	}
	project := doc.Project
	if project == "" && doc.Scope != "reference" {
		project = defaultProject
	}
	if project == "" && doc.Scope != "reference" && (doc.Category == gyrus.CategoryProduct || doc.Category == gyrus.CategoryBusinessLogic) {
		project = "gyrus"
	}
	if project != "" {
		if err := ValidateSegment("project", project); err != nil {
			return "", err
		}
		return path.Join("concepts", doc.Tenant, "projects", project, doc.ID+".md"), nil
	}
	return path.Join("concepts", doc.Tenant, "reference", doc.ID+".md"), nil
}

// SetLocation derives scope and project from an anchored collection path.
func SetLocation(doc *gyrus.Concept, locator string) error {
	doc.FilePath = locator
	clean := strings.ReplaceAll(locator, "\\", "/")
	parts := strings.Split(clean, "/")
	var name string
	if n := len(parts); n >= 4 && parts[n-4] == "concepts" && parts[n-2] == "reference" {
		name = strings.Join(parts[n-3:], "/")
	} else if n >= 5 && parts[n-5] == "concepts" && parts[n-3] == "projects" {
		name = strings.Join(parts[n-4:], "/")
	} else {
		return &ValidationError{Field: "location", Message: "expected canonical concepts collection path"}
	}
	extension := path.Ext(name)
	if extension != ".md" && extension != ".json" {
		return &ValidationError{Field: "location", Message: "expected Markdown or JSON concept"}
	}
	name = strings.TrimSuffix(name, extension)
	tenant, project, id, err := ParseName(name)
	if err != nil {
		return err
	}
	if doc.ID != id || doc.Tenant != tenant {
		return &ValidationError{Field: "location", Message: "filename or tenant does not match concept metadata"}
	}
	doc.Name, doc.Project, doc.Scope = name, project, "reference"
	if project != "" {
		doc.Scope = "project"
	}
	return nil
}

// Visit validates a detached collection snapshot before invoking its consumer.
func Visit(ctx context.Context, docs []gyrus.Concept, report gyrus.SyncReport, visit func(gyrus.Concept) error) (gyrus.SyncReport, error) {
	seen := make(map[string]string)
	for _, doc := range docs {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		if err := Validate(&doc); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", doc.FilePath, err))
			continue
		}
		if previous, exists := seen[doc.Name]; exists {
			report.Errors = append(report.Errors, fmt.Sprintf("duplicate concept ID %q at %s and %s", doc.Name, previous, doc.FilePath))
			continue
		}
		seen[doc.Name] = doc.FilePath
		if err := visit(doc); err != nil {
			return report, err
		}
		report.IndexedFiles++
	}
	return report, nil
}
