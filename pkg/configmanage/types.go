package configmanage

type ManagedResource struct {
	Host     string                 `yaml:"host" json:"host"`
	Password string                 `yaml:"password" json:"password"`
	Files    []FileSpecification    `yaml:"files" json:"files"`
	Packages []PackageSpecification `yaml:"packages" json:"packages"`
	Command  []string               `yaml:"command" json:"command"`
}

type FileSpecification struct {
	Name string `yaml:"name" json:"name"`
	Path string `yaml:"path" json:"path"`
	Mode string `yaml:"mode" json:"mode"`
}

type PackageSpecification struct {
	Package string `yaml:"package" json:"package"`
	Version string `yaml:"version" json:"version"`
}

// These structs describe actions that can be taken on resources

type FileResourceDiff struct {
	Operation    string
	Target       string
	UpdateValue  interface{}
	FileResource FileSpecification
}

type PackageResourceDiff struct {
	Operation       string
	PackageResource PackageSpecification
}
