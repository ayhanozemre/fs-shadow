package connector

type Path interface {
	IsVirtual() bool
	IsDir() bool
	Exists() bool
	Name() string
	String() string
	ParentPath() Path
	ExcludePath(p Path) Path
}
