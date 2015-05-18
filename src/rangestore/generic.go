package rangestore

// generic interface for any Store
type Store interface {
	// lookup cluster
	ClusterLookup(string) (*[]string, error)     // cluster
	KeyLookup(string, string) (*[]string, error) // cluster and key

	// lookup reverse
	KeyReverseLookup(string) (*[]string, error)                     // just a reverse lookup on a node
	KeyReverseLookupAttr(string, string) (*[]string, error)         // reverse lookup where value and key are passed
	KeyReverseLookupHint(string, string, string) (*[]string, error) // reverse lookup where value and key are passed with an hint
}
