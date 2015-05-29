package rangeexpr

import "testing"

// this is to avoid compiler optimization, used in benchmarking
var result *[]string

//////////////////
// Benchmarking //
//////////////////

// expand the query
func expandQuery(q string) (*[]string, error) {
	var r = &RangeExpr{Buffer: q}
	r.Init()
	r.Expression.Init(q)
	err := r.Parse()
	if err != nil {
		return &[]string{}, err
	}
	r.Execute()
	result, errs := r.Evaluate(store)
	if len(errs) != 0 {
		return &[]string{}, err
	}

	return result, nil
}

func benchMark(q string, b *testing.B) {
	var r *[]string
	for n := 0; n < b.N; n++ {
		r, _ = expandQuery(q)
	}
	// this is to avoid compiler optimizations
	result = r
}

func BenchmarkCluster(b *testing.B) {
	benchMark("%ops", b)
}

func BenchmarkKeyLookup(b *testing.B) {
	benchMark("%ops-prod-vpc1-range:AUTHORS", b)
}

func BenchmarkKeyReverse(b *testing.B) {
	benchMark("*range1001.ops.example.com", b)
}

func BenchmarkKeyReverseAttr(b *testing.B) {
	benchMark("Ops;AUTHORS", b)
}

func BenchmarkKeyReverseHint(b *testing.B) {
	benchMark("Ops;AUTHORS:ops", b)
}
