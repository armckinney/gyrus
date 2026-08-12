package format

import (
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// StorageFormat abstracts document serialization and deserialization for storage engines.
type StorageFormat interface {
	Serialize(doc *gyrus.Document) ([]byte, error)
	Deserialize(data []byte) (*gyrus.Document, error)
}
