package cosmos

type CosmosPackage interface {
	GetName() string
	GetVersion() string
	GetDescription() string
	GetConfig() map[string]interface{}
}

type CosmosRepository interface {
	FindAllPackageVersions(name string) ([]CosmosPackage, error)
	FindLatestPackageVersion(name string) (CosmosPackage, error)
	FindPackageVersion(name string, version string) (CosmosPackage, error)
}
