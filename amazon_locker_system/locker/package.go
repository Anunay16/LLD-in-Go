package locker

type PackageSize string

const (
	LargePackageSize  PackageSize = "large"
	MediumPackageSize PackageSize = "medium"
	SmallPackageSize  PackageSize = "small"
)

type Package struct {
	id      string
	size    PackageSize
	agentId string
}

func NewPackage(id, agentId string, size PackageSize) *Package {
	return &Package{
		id:      id,
		agentId: agentId,
		size:    size,
	}
}

func (pkg *Package) GetID() string {
	return pkg.id
}

func (pkg *Package) GetSize() PackageSize {
	return pkg.size
}
