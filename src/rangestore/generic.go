package rangestore

// a generic store so all the other stores can be type-casted to this
// to the generic store as follows
// var store rangestore.Store     /* create an interface */
// store = s.(rangestore.Store)   /* type-cast to generic store interface */

// generic interface for any Store
type Store interface {
	// lookup cluster
	ClusterLookup(*[]string) (*[]string, error)     // cluster
	KeyLookup(*[]string, string) (*[]string, error) // cluster and key

	// lookup reverse
	KeyReverseLookup(string) (*[]string, error)                     // just a reverse lookup on a node
	KeyReverseLookupAttr(string, string) (*[]string, error)         // reverse lookup where value and key are passed
	KeyReverseLookupHint(string, string, string) (*[]string, error) // reverse lookup where value and key are passed with an hint
}

//////////////////////
// Generic Template //
//////////////////////

type GenericStore struct {
	// Some attrs
}

// initialize the store
func ConnectGenericStore(dir string) (*GenericStore, error) {
	return nil, nil
}

////////////////////
// LOOKUP CLUSTER //
////////////////////

func (f *GenericStore) ClusterLookup(cluster *[]string) (*[]string, error) {
	return &[]string{}, nil
}

func (f *GenericStore) KeyLookup(cluster *[]string, key string) (*[]string, error) {
	return &[]string{}, nil
}

////////////////////
// LOOKUP REVERSE //
////////////////////

func (f *GenericStore) KeyReverseLookup(key string) (*[]string, error) {
	return &[]string{}, nil
}

func (f *GenericStore) KeyReverseLookupAttr(key string, attr string) (*[]string, error) {
	return &[]string{}, nil
}

func (f *GenericStore) KeyReverseLookupHint(key string, attr string, hint string) (*[]string, error) {
	return &[]string{}, nil
}
