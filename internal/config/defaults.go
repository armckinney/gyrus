package config

const (
	// DefaultStorageProvider is the default document persistence driver.
	DefaultStorageProvider = "localfs"

	// DefaultIndexProvider is the default metadata index driver.
	DefaultIndexProvider = "sqlite"

	// DefaultSearchProvider is the default search driver.
	DefaultSearchProvider = "sqlite"

	// DefaultGraphProvider is the default graph relation driver.
	DefaultGraphProvider = "sqlite"

	// DefaultOwnerGroup is the default OKF document owner group.
	DefaultOwnerGroup = "root"

	// DefaultStorageRoot is the default directory relative to workspace or home.
	DefaultStorageRoot = ".gyrus"

	// GlobalConfigFileName is the canonical configuration file name.
	GlobalConfigFileName = ".gyrus.yaml"
)

// ValidStorageProviders lists all recognized storage provider names.
var ValidStorageProviders = []string{
	"localfs", "git", "s3", "aws_s3", "azure", "azure_blob",
	"gcs", "gcp_gcs", "gcp", "blob", "postgres",
}

// ValidSearchProviders lists all recognized search provider names.
var ValidSearchProviders = []string{
	"sqlite", "sqlite_fts5", "postgres_fts", "vector",
}

// ValidIndexProviders lists all recognized index provider names.
var ValidIndexProviders = []string{
	"sqlite", "postgres",
}

// ValidGraphProviders lists all recognized graph provider names.
var ValidGraphProviders = []string{
	"sqlite", "postgres",
}
