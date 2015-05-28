// Test all the possible range expressions and query against
// the teststore. Reason for creating each function for a test
// case was to ease the debugging process and I wasn't sure whether
// I wanted to piecemeal the cases.
// FIXME: Create a map with key as query and value as the expected result
//        and use this map to test the queries.

package rangeexpr

import (
	"rangestore"
	"testing"
)

var store, err = rangestore.ConnectTestStore("Test Store") // this can never return error

// ""
// empty string
func TestParsing00(t *testing.T) {
	var q = ""
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [empty string is valid]", q)
	}
}

// "1"
// starts with a numeric
func TestParsing01(t *testing.T) {
	var q = "1"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err == nil {
		t.Errorf("Expected Error, (Query: %s) should NOT BE parsed [starts with 0-9]", q)
	}
}

// "a"
// starts with a alphabet
func TestParsing02(t *testing.T) {
	var q = "a"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [starts with a-z]", q)
	}
	r.Execute()
	result, errs := r.Evaluate(store)
	if len(errs) != 0 || !compare(*result, []string{"a"}) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", []string{"ops"}, *result)
	}
}

// "%ops"
// starts with a alphabet
func TestParsing03(t *testing.T) {
	var q = "%ops"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [starts with %%[a-z]]", q)
	}
	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "%aa1-b-c-d1"
// valid cluster string, but contains %
func TestParsing04(t *testing.T) {
	var q = "%aa1-b-c-d1"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [is of %%[a-z][a-z0-9-_]*", q)
	}
}

// "%a-b-c%d"
// invalid cluster string
func TestParsing05(t *testing.T) {
	var q = "%a-b-c%d"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err == nil {
		t.Errorf("Expected Error, (Query: %s) should NOT BE parsed [contains %%, where not expected]", q)
	}
}

// "%a-b-c-1
// valid cluster string, ends with -1
func TestParsing06(t *testing.T) {
	var q = "%a-b-c-1"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [ends with -1]", q)
	}
}

// "%a-b-%d"
// invalid ends with -%d
func TestParsing07(t *testing.T) {
	var q = "%a-b-%d"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err == nil {
		t.Errorf("Expected Error, (Query: %s) should NOT BE parsed [Unexpected %% (before d)]", q)
	}
}

// "%RANGE"
// Top Level
func TestParsing08(t *testing.T) {
	var q = "%RANGE"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Top Level Lookup, %, %% etc]", q)
	}
	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops", "data"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "%%RANGE"
// valid intersection operator query
func TestParsing09(t *testing.T) {
	var q = "%%RANGE"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Top Level Lookup, %, %% etc]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod", "data-prod", "data-qa"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

/* COMBINED EXPRESSIONS */

// "%a-b,%d"
// valid union operator query
func TestParsingComb01(t *testing.T) {
	var q = "%a-b,%d"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (union), eg %%foo,%%bar]", q)
	}
}

// "%a-b,%d,"
// invalid union operator query ending with ,
func TestParsingComb02(t *testing.T) {
	var q = "%a-b,%d,"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err == nil {
		t.Errorf("Expected Error, (Query: %s) should NOT BE parsed [Combined Expression (union) should not end with ,]", q)
	}
}

// "%a-b,&%d"
// valid intersection operator query
func TestParsingComb03(t *testing.T) {
	var q = "(%ops-prod ,& %data-prod), test"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (intersection), eg %%foo,&%%bar]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"test"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "%a-b,-%d"
// valid difference operator query
func TestParsingComb04(t *testing.T) {
	var q = "%ops ,- %data"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (difference), eg %%foo,-%%bar]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "%data-prod ,- %ops-prod ,- vpc2"
// valid difference operator query
func TestParsingComb05(t *testing.T) {
	var q = "%data-prod ,- %ops-prod ,- vpc2"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (difference), eg %%foo,-%%bar]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"data-prod-vpc1", "data-prod-vpc2", "data-prod-vpc3"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "(%date-prod ,- %ops-prod) ,- vpc2"
// difference with grouping together
func TestParsingComb06(t *testing.T) {
	var q = "(%data-prod ,- %ops-prod) ,- data-prod-vpc2"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (union, difference), eg (%%foo,-%%bar),%%cow]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"data-prod-vpc1", "data-prod-vpc3"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "%ops-prod, (%data-prod ,- vpc2)"
// union and difference together
func TestParsingComb07(t *testing.T) {
	var q = "%ops-prod, (%data-prod ,- data-prod-vpc2)"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (union, difference), eg (%%foo,-%%bar),%%cow]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod-vpc1", "ops-prod-vpc2", "data-prod-vpc1", "data-prod-vpc3"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "(%ops-prod ,& vpc2) , %data-prod ,- vpc2"
// union, difference and intersection together
func TestParsingComb08(t *testing.T) {
	var q = "(%ops-prod ,& vpc2) , %data-prod ,- %data-qa"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (union, difference), eg (%%foo,-%%bar),%%cow]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"data-prod-vpc1", "data-prod-vpc2", "data-prod-vpc3"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "%ops-prod ,& (vpc2 , %data-prod) ,- vpc2"
// union, difference and intersection together
func TestParsingComb09(t *testing.T) {
	var q = "%ops-prod ,& (ops-prod-vpc2 , %data-prod) ,- %data-qa"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (union, difference), eg (%%foo,-%%bar),%%cow]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod-vpc2"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "%%data-qa-vpc5-log:QAFOR"
// union, difference and intersection together
func TestParsingComb10(t *testing.T) {
	var q = "%%data-qa-vpc5-log:QAFOR ,& %data"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (union, difference), eg (%%foo,-%%bar),%%cow]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"data-prod", "data-qa"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

/* KEYS */

// "%ops-prod-vpc1-range:AUTHORS"
// upper case key
func TestParsingKey01(t *testing.T) {
	var q = "%ops-prod-vpc1-range:AUTHORS"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [ Keys, eg (%%foo:KEY)]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"Vigith Maurice"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "%aa-bb-cc:key"
// lower case key
func TestParsingKey02(t *testing.T) {
	var q = "%aa-bb-cc:key"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err == nil {
		t.Errorf("Expected Error, (Query: %s) should NOT BE parsed [ Keys should be always UPPERCASE, eg (%%foo:BAR)]", q)
	}
}

// "(*Ops;AUTHORS , *Vigith Maurice;AUTHORS) ,& (ops-prod-vpc1-range, ops-prod-vpc1-mon)"
// (%ops-prod-vpc1-range:AUTHORS"
// upper case key
func TestParsingKey03(t *testing.T) {
	var q = "(*Ops;AUTHORS , *Vigith Maurice;AUTHORS) ,& (ops-prod-vpc1-range, ops-prod-vpc1-mon)"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [ Keys, eg (%%foo:KEY)]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod-vpc1-range", "ops-prod-vpc1-mon"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// Reverse Lookup

// "*range1001.ops.example.com"
// Reverse Lookup
func TestRevParsing01(t *testing.T) {
	var q = "*range1001.ops.example.com"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [basic reverse lookup]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod-vpc1-range"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "*Vigith Maurice;AUTHORS"
// Reverse Lookup with Attr
func TestRevParsing02(t *testing.T) {
	var q = "*Vigith Maurice;AUTHORS"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [reverse lookup with attr]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod-vpc1-range"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "*Ops;AUTHORS:ops-prod-vpc1-mon"
// Reverse Lookup with Attr and Hint
func TestRevParsing03(t *testing.T) {
	var q = "*Ops;AUTHORS:ops-prod-vpc1-mon"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [reverse lookup with attr and hint]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod-vpc1-mon", "ops-prod-vpc2-mon"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "(*Ops;AUTHORS , *Vigith Maurice;AUTHORS) ,& (ops-prod-vpc1-range, ops-prod-vpc1-mon)"
func TestRevParsing04(t *testing.T) {
	var q = "(*Ops;AUTHORS , *Vigith Maurice;AUTHORS) ,& (ops-prod-vpc1-range, ops-prod-vpc1-mon)"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [reverse lookup with attr and hint]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod-vpc1-mon", "ops-prod-vpc1-range"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// "*Ops;AUTHORS , (*Vigith Maurice;AUTHORS ,& ops-prod-vpc1-range) , ops-prod-vpc1-mon"
func TestRevParsing05(t *testing.T) {
	var q = "*Ops;AUTHORS , (*Vigith Maurice;AUTHORS ,& ops-prod-vpc1-range) , ops-prod-vpc1-mon"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [reverse lookup with attr and hint]", q)
	}

	r.Execute()
	result, errs := r.Evaluate(store)
	var expected = []string{"ops-prod-vpc1-mon", "ops-prod-vpc1-range", "ops-prod-vpc2-mon"}
	if len(errs) != 0 || !compare(*result, expected) {
		t.Errorf("Expected NO Evaluate Error, (Query: %s) should BE %s [Got: %s]", q, expected, *result)
	}
}

// Internal Function

// Compare 2 Arrays, items need not be in correct order
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
