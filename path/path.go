package connector

type Params struct {
	From Path
	To   Path
}

type Path interface {
	IsVirtual() bool
	IsDir() bool
	Exists() bool
	Name() string
	String() string
	ParentPath() Path
	ExcludePath(p Path) Path
	Info() *FileInfo
}

type FileInfo struct {
	IsDir      bool
	Size       int64
	CreatedAt  int64
	Permission string
}
