package repomapper

// Symbol represents a code definition (function, struct, class, etc.)
type Symbol struct {
	Name      string
	Kind      string
	Signature string
	Line      int
}

// FileInfo holds symbol and ranking data for a file
type FileInfo struct {
	Path    string
	Symbols []Symbol
	Rank    int
}

// RepoMapper generates a skeleton map of a codebase
type RepoMapper struct {
	impl repoMapperImpl
}

// repoMapperImpl is the platform-specific implementation interface
type repoMapperImpl interface {
	GenerateMap(root string) (string, error)
}

// NewRepoMapper initializes a new RepoMapper with supported languages and queries
func NewRepoMapper() *RepoMapper {
	return &RepoMapper{
		impl: newRepoMapperImpl(),
	}
}

// GenerateMap scans the root directory and produces a hierarchical text tree map
func (rm *RepoMapper) GenerateMap(root string) (string, error) {
	return rm.impl.GenerateMap(root)
}
