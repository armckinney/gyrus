package format

import (
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// StorageFormat abstracts concept serialization and deserialization for storage engines.
type StorageFormat interface {
	Serialize(doc *gyrus.Concept) ([]byte, error)
	Deserialize(data []byte) (*gyrus.Concept, error)
}
