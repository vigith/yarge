package rangeexpr

import (
	"testing"
)

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
}

// "%a"
// starts with a alphabet
func TestParsing03(t *testing.T) {
	var q = "%a"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [starts with %%[a-z]]", q)
	}
}

// "%a-b-c-d"
// valid cluster string, but contains %
func TestParsing04(t *testing.T) {
	var q = "%a-d-c"
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
// valid intersection operator query
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

// "%"
// valid intersection operator query
func TestParsing08(t *testing.T) {
	var q = "%"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Top Level Lookup, %, %% etc]", q)
	}
}

// "%%"
// valid intersection operator query
func TestParsing09(t *testing.T) {
	var q = "%%"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Top Level Lookup, %, %% etc]", q)
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

// "%a-b&%d"
// valid intersection operator query
func TestParsingComb03(t *testing.T) {
	var q = "%ab-cd&%d"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (intersect), eg %%foo&%%bar]", q)
	}
}

// "%a-b^%d"
// valid difference operator query
func TestParsingComb04(t *testing.T) {
	var q = "%a1d^d"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (difference), eg %%foo^%%bar]", q)
	}
}

// "%a,(%c^%dd-ee)"
// union and difference together
func TestParsingComb05(t *testing.T) {
	var q = "%a,(%c^%dd-ee)"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [Combined Expression (union, difference), eg (%%foo^%%bar),%%cow]", q)
	}
}

/* KEYS, etc */

// "%aa-bb-cc:DD"
// union and difference together
func TestParsingMisc01(t *testing.T) {
	var q = "%aa-bb-cc:DD"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		t.Errorf("Expected NO Error, (Query: %s) should BE parsed [ Keys, eg (%%foo:KEY)]", q)
	}
}

// "%aa-bb-cc:DD"
// union and difference together
func TestParsingMisc02(t *testing.T) {
	var q = "%aa-bb-cc:key"
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err == nil {
		t.Errorf("Expected Error, (Query: %s) should NOT BE parsed [ Keys should be always UPPERCASE, eg (%%foo:BAR)]", q)
	}
}
