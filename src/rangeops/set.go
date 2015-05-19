package rangeops

// union of 2 sets and populates the result set
// this works in worst case O(N x M) complexity
// ASSUMPTION: Arguments should be a set, (definition
// of set says "no duplicates")
func Union(set1 *[]string, set2 *[]string, r *[]string) {
	// * find the larger and shorter of the 2 sets
	// * append the larger one to the result
	// * iterate over the shorter set and append
	//   if that element is not in the larger set
	var longer *[]string
	var shorter *[]string

	// find who is the larger and shorter, then
	// set the pointers	accordingly
	if len(*set1) >= len(*set2) {
		longer = set1
		shorter = set2
	} else {
		longer = set2
		shorter = set1
	}

	// append the longer
	*r = append(*r, *longer...)

	// insert if shorter not present in larger
	for _, elem := range *shorter {
		if !contains(longer, elem) {
			*r = append(*r, elem)
		}
	}

	return
}

// intersection of 2 sets and populates the result set
// this works in worst case O(N x M) complexity
// ASSUMPTION: Arguments should be sets, (definition
// of set says "no duplicates")
func Intersection(set1 *[]string, set2 *[]string, r *[]string) {
	// * find the larger and shorter of the 2 sets
	// * iterate over the shorter set and append
	//   if that element is in the larger set
	var longer *[]string
	var shorter *[]string

	// find who is the larger and shorter, then
	// set the pointers	accordingly
	if len(*set1) >= len(*set2) {
		longer = set1
		shorter = set2
	} else {
		longer = set2
		shorter = set1
	}

	// insert if shorter is present in larger
	for _, elem := range *shorter {
		if contains(longer, elem) {
			*r = append(*r, elem)
		}
	}

	return
}

// difference of 2 sets and populates the result set
// this works in worst case O(N x M) complexity
// ASSUMPTION: Arguments should be sets, (definition
// of set says "no duplicates")
// eg, A = {1,2,3}
//     B = {2,3,4}
//     A - B = {1}
// ie, A - B = { x <- A | x !<- B }
func Difference(set1 *[]string, set2 *[]string, r *[]string) {
	// * find the larger and shorter of the 2 sets
	// * iterate over the larger set and append
	//   if that element is not in the shorter set
	var longer *[]string
	var shorter *[]string

	// find who is the larger and shorter, then
	// set the pointers	accordingly
	if len(*set1) >= len(*set2) {
		longer = set1
		shorter = set2
	} else {
		longer = set2
		shorter = set1
	}

	// insert if larger not present in shorter
	for _, elem := range *longer {
		if !contains(shorter, elem) {
			*r = append(*r, elem)
		}
	}

	return
}

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

///////////////////////
// Internal Fuctions //
///////////////////////

// tests whether an element is present in a array,
// returns true if found, else false
func contains(arr1 *[]string, elem string) bool {
	for _, i := range *arr1 {
		if elem == i {
			return true
		}
	}
	return false
}
