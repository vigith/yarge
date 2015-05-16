package rangestore

// FileStore
// When using FileStore, we will be using Files
// as the store.

// type FileStore struct {
// 	StorePath string
// }

// // TODO: check whether the StorePath Exists, etc
// func ConnectFStore(path string) (f FileStore, err error) {
// 	f = FileStore{StorePath: path}
// 	return f, nil
// }

// // function to connect to store
// func (f FileStore) ClusterLookup(string) []string {
// 	return nil
// }

// given an array, convert it into a set in place ie, remove duplicates from array
// NOTE: Not the best algorithm, this is not space and time efficient.
//       This function is written only for FileStore, because people make
//       mistakes, not programs.
func ArrayToSet(array *[]string) {
	// map to store the
	m := map[string]bool{}
	// replace the current array in place such that, if we
	// have not seen the element, insert it, But, insert such that
	// insert at the position in the array where index = len(m) (HINT:
	// m has unique elements of the array)
	for _, v := range *array {
		if _, seen := m[v]; !seen {
			(*array)[len(m)] = v
			m[v] = true
		}
	}
	// re-slice s to the number of unique values
	*array = (*array)[:len(m)]

	return
}
