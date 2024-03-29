package service

// CatalogConfig represents details re: which modules' services should be
// included or excluded from the catalog
type CatalogConfig struct {
	EnableMigrationServices bool
	EnableDRServices        bool
}

type tempCatalogConfig struct {
	CatalogConfig
	EnableMigrationServicesStr string
	EnableDRServicesStr        string
}

// NewCatalogConfigWithDefaults returns a CatalogConfig object with default
// values already applied. Callers are then free to set custom values for the
// remaining fields and/or override default values.
func NewCatalogConfigWithDefaults() CatalogConfig {
	return CatalogConfig{
		EnableMigrationServices: false,
	}
}
