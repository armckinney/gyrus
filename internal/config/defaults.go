package config

const (
	// DefaultStorageProvider is the default concept persistence driver.
	DefaultStorageProvider = "localfs"

	// DefaultIndexProvider is the default metadata index driver.
	DefaultIndexProvider = "okf"

	// DefaultSearchProvider is the default search driver.
	DefaultSearchProvider = "okf"

	// DefaultGraphProvider is the default graph relation driver.
	DefaultGraphProvider = "okf"

	// DefaultTenant is the default OKF concept tenant.
	DefaultTenant = "root"

	// DefaultProject is the default OKF concept project.
	DefaultProject = "default"

	// DefaultStorageRoot is the default directory relative to project or home.
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
	"okf",
	"postgres_fts", "vector",
}

// ValidIndexProviders lists all recognized index provider names.
var ValidIndexProviders = []string{
	"okf",
	"postgres",
}

// ValidGraphProviders lists all recognized graph provider names.
var ValidGraphProviders = []string{
	"okf",
	"postgres",
}
