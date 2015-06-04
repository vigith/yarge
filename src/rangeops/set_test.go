package rangeops

import "testing"

// few functions for testing

// Compare 2 Arrays, items need not be in correct
// order
func compare(arr1, arr2 []string) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	if len(arr1) == 0 {
		return true
	}

	var flag bool
	for _, value1 := range arr1 {
		for _, value2 := range arr2 {
			if value1 == value2 {
				flag = true
			}
		}
		if !flag {
			return false
		}
	}

	return true
}

// test contains
func Test_contain01(t *testing.T) {
	var arr1 = []string{"foo", "bar", "moo", "moo"}

	// test for a true cond
	var elem1 = "foo"
	var boolean bool
	boolean = contains(&arr1, elem1)
	if !boolean {
		t.Errorf("Expected NO Error, [%s] is present in array %s", elem1, arr1)
	}

	// test for false condition
	var elem2 = "notthere"
	boolean = contains(&arr1, elem2)
	if boolean {
		t.Errorf("Expected Error, [%s] is NOT present in array %s", elem2, arr1)
	}
}

// test union
func TestUnion01(t *testing.T) {
	var set1 = []string{}
	var set2 = []string{}

	var res = make([]string, 0)

	Union(&set1, &set2, &res)
	if !compare(res, []string{}) {
		t.Errorf("Expected NO Error, both sets are empty, so Union should return empty set, set1 %s set2 %s : result %s", set1, set2, res)
	}

	set1 = []string{"foo", "bar", "moo"}
	set2 = []string{"bar"}
	res = make([]string, 0)
	Union(&set1, &set2, &res)
	if !compare(res, set1) {
		t.Errorf("Expected NO Error, Union should return set1, because set2 is a subset of set1, set1 %s set2 %s : result %s", set1, set2, res)
	}

	set1 = []string{"foo", "bar", "moo"}
	set2 = []string{"cow"}
	res = make([]string, 0)
	Union(&set1, &set2, &res)
	if compare(res, set1) {
		t.Errorf("Expected Error, Union should return set1 + set2, because set2 is not a subset of set1, set1 %s set2 %s : result %s", set1, set2, res)
	}
}

// test intersection
func TestIntersection01(t *testing.T) {
	var set1 = []string{}
	var set2 = []string{}

	var res = make([]string, 0)

	Intersection(&set1, &set2, &res)
	if !compare(res, []string{}) {
		t.Errorf("Expected NO Error, both sets are empty, so Intersection should return empty set, set1 %s set2 %s : result %s", set1, set2, res)
	}

	set1 = []string{"foo", "bar", "moo"}
	set2 = []string{"bar"}
	res = make([]string, 0)
	Intersection(&set1, &set2, &res)
	if !compare(res, set2) {
		t.Errorf("Expected NO Error, Intersection should return set2, because set2 is a proper subset of set1, set1 %s set2 %s : result %s", set1, set2, res)
	}

	set1 = []string{"foo", "bar", "moo"}
	set2 = []string{"cow"}
	res = make([]string, 0)
	Intersection(&set1, &set2, &res)
	if !compare(res, []string{}) {
		t.Errorf("Expected NO Error, Intersection should return empty set, because set2 and set1 are mutuall exclusive, set1 %s set2 %s : result %s", set1, set2, res)
	}
}

// test difference
func TestDifference01(t *testing.T) {
	var set1 = []string{}
	var set2 = []string{}

	var res = make([]string, 0)

	Difference(&set1, &set2, &res)
	if !compare(res, []string{}) {
		t.Errorf("Expected NO Error, both sets are empty, so Difference should return empty set, set1 %s set2 %s : result %s", set1, set2, res)
	}

	set1 = []string{"bar", "cow"}
	set2 = []string{"foo", "bar", "moo"}
	res = make([]string, 0)
	Difference(&set1, &set2, &res)
	if !compare(res, []string{"foo", "moo"}) {
		t.Errorf("Expected NO Error, Difference should return set1 - set2, set1 %s set2 %s : result %s", set1, set2, res)
	}

	set1 = []string{"foo", "bar", "moo"}
	set2 = []string{"cow"}
	res = make([]string, 0)
	Difference(&set1, &set2, &res)
	if !compare(res, set1) {
		t.Errorf("Expected NO Error, Difference should return set1, because set2 and set1 are mutuall exclusive, set1 %s set2 %s : result %s", set1, set2, res)
	}
}

// Array to Set
func TestArrayToSet01(t *testing.T) {
	var set1 = []string{}
	ArrayToSet(&set1)
	if !compare(set1, []string{}) {
		t.Errorf("Expected NO Error, set is empty, so ArrayToSet should return empty set, set1 %s : result %s", set1, []string{})
	}

	set1 = []string{"foo", "bar", "moo"}
	ArrayToSet(&set1)
	if !compare(set1, []string{"foo", "moo", "bar"}) {
		t.Errorf("Expected NO Error, ArrayToSet should return set1 after removing duplicates, set1 %s : result %s", set1, []string{"foo", "moo", "bar"})
	}

	set1 = []string{"foo", "foo", "bar", "moo", "bar"}
	ArrayToSet(&set1)
	if !compare(set1, []string{"foo", "moo", "bar"}) {
		t.Errorf("Expected NO Error, ArrayToSet should return set1 after removing duplicates, set1 %s : result %s", set1, []string{"foo", "moo", "bar"})
	}
}
