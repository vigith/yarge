// AST Processor for Range Expression
// Author: Vigith Maurice

package rangeexpr

import (
	"rangeops"
	"rangestore"
)

type Type uint8

const (
	typeData Type = iota
	// set operations
	typeUnion
	typeIntersection
	typeDifference
	// cluster lookup
	typeClusterLookup
	typeKeyLookup // 5
	// reverse lookup
	typeKeyReverseLookup
	typeKeyReverseLookupAttr
	typeKeyReverseLookupHint
)

// each token will be represented as a bytecode
type ByteCode struct {
	T     Type
	Value string
}

// each expression once parsed will be represented an array
// of bytecodes and the number of bytecodes in the array
type Expression struct {
	Code []ByteCode
	Top  int
}

// create a slice to hold the expression as bytecodes
func (e *Expression) Init(expression string) {
	e.Code = make([]ByteCode, len(expression))
}

// add an operator on to the expression array
func (e *Expression) addOperator(operator Type) {
	code, top := e.Code, e.Top
	e.Top++
	code[top].T = operator
}

// add a value on to the expression array
// TODO: we can remove unwanted stuff in here
// eg, remove trailing spaces in reverse lookup
func (e *Expression) addValue(value string) {
	code, top := e.Code, e.Top
	e.Top++
	code[top].Value = value
}

// Accepts the interface for connection to store.
// Returns a pointer to array of strings (result) and error
func (e *Expression) Evaluate(s interface{}) (*[]string, []error) {
	// simplest case, no ByteCode because expr was an empty string
	if len(e.Code) == 0 {
		return nil, nil
	}

	// typecast the store (s) to the generic store
	var store rangestore.Store
	store = s.(rangestore.Store)

	// Create an array of errors
	var errs = make([]error, 0)

	// create a stack to hold pointers to string, and a index
	// to track the stack. stack pointers will undergo inplace
	// modifications
	stack, top := make([]*[]string, len(e.Code)), 0 // array of point to array of string
	// Rules:
	// =====
	// Rule 0: Follow the my basic rules, we will get our AST processed in
	//         1-pass :-).
	// Rule 1: As soon as we hit a cluster set operation, take the
	//         top 2 elements of the stack the replace the top most
	//         with the set operation result. [Binary Operators]
	// Rule 2: If it is not a cluster set operation, try to Peek Forward
	//         Sensibly to make more sense of the input stream, if no
	//         sense is made, push on to the stack. [Unary Operators]
	//         Denition of Peek Forward Sensibly is as follows:
	//         a. If it is a ClusterLookup, try to see whether it
	//            is a lookup with key involved (ie, KeyLookup)
	//         b. If it is a KeyReverseLookup, try to check whether
	//            we have an attr (ReverseLookupAttr) and or a hint
	//            (KeyReverseLookupHint)
	// Rule 3: stack[0] is always the result
	var ptr = 0 // ptr to iterate over e.Code
	// we will have to manually iterate the e.Code,
	// this is because we might have to jump ahead due to
	// our lookahead cases
	for ptr < e.Top {
		code := e.Code[ptr]
		switch code.T {
		// Data Insertion
		// if it not an operation, but just data
		case typeData:
			// convert the data as an array string
			stack[top] = &[]string{code.Value}
			top++

		// Cluster Lookup
		// --------------
		// if type == ClusterLookup, peek 2 ahead to check whether
		// it is a KeyLookup, if yes continue and let KeyLookup take care
		// of expanding, else expand in place
		// if type == KeyLookup, take the next elem from the stack and
		// do an inplace expand.
		// eg,
		// Case 1:
		//   %d1 => [d1, ClusterLookup]
		//   stack => [ nil, [d1,], ] <= push d1
		//   stack => [ nil, lookup([d1,]), ] <= ClusterLookup (inplace)
		// Case 2:
		//   d2 => [d2]
		//   stack => [nil, [d2,] ] <= push d2
		// Case 3:
		//   %d1:D2 => [d1, ClusterLookup, D2, KeyLookup]
		//   stack => [nil, [d1,] ] <= push d1
		//   stack => [nil, [d1,] ] <= (at ClusterLookup) peeks +2 and sees KeyLookup so passes
		//   stack => [nil, [d1,], [d2,] ] <= push d2
		//   stack => [nil, lookup([d1,], d2) ] <= (at KeyLookup) it takes the next value and lookups with key

		case typeClusterLookup:
			// peek first, if true continue
			if ptr < e.Top-2 && e.Code[ptr+2].T == typeKeyLookup {
				ptr++
				continue
			}
			result, err := store.ClusterLookup(stack[top-1])
			// store the addr of the result
			stack[top-1] = result
			// append the errors
			if err != nil {
				errs = append(errs, err)
			}

		case typeKeyLookup:
			result, err := store.KeyLookup(stack[top-2], (*stack[top-1])[0])
			// store the addr of the result
			// we will have to de-dup if stack[top-2] has more than 1 element
			// 'coz there could intersecting elements for KEY in different clusters
			if len(*(stack[top-2])) == 1 {
				stack[top-2] = result
			} else {
				// conver to a set
				rangeops.ArrayToSet(result)
				stack[top-2] = result
			}
			// reset top to nil, that value is no more useful to us
			stack[top-1] = nil
			// binary operator, pops of 2 elements to produce a result
			top--
			// append the errors
			if err != nil {
				errs = append(errs, err)
			}

		// Reverse Lookup
		// --------------
		// if type == ReverseKeyLookup, peek 2 ahead to check whether it is a
		// ReverseKeyLookupAttr, if yes continue and let ReverseKeyLookupAttr
		// take care of expanding, else expand in place
		// if type == ReverseKeyLookupAttr, peek 2 ahead to check whether it is a
		// ReverseLookupHint, if yes continue and let ReverseLookupHint take care
		// of expanding
		// if type == ReverseKeyLookupHint, take the values from the stack and
		// do an inplace expand
		// eg,
		// Case 1:
		//   *d1 => [d1, ReverseKeyLookup]
		//   stack => [ nil, [d1,], ] <= push d1
		//   stack => [ nil, reverse([d1,]), ] <= ReverseKeyLookup (inplace)
		// Case 2:
		//   *d1;A => [d1, ReverseKeyLookup, A, ReverseKeyLookupAttr]
		//   stack => [nil, reverse([d1,], A) ] <=
		// Case 3:
		//   %d1;A:d2 => [d1, ReverseKeyLookup, A, ReverseKeyLookupAttr, d2, ReverseKeyLookupHint]
		//   stack => [nil, [d1,] ] <= push d1
		//   stack => [nil, [d1,] ] <= (at ReverseKeyLookup) peeks and sees ReverseKeyLookupAttr so passes
		//   stack => [nil, [d1,], A, ] <= (at ReverseKeyLookupAttr) peeks and sees ReverseKeyLookupHint so passes
		//   stack => [nil, reverse([d1,], A, d2) ] <= (at ReverseKeyLookupHint) takes the 2 vales from stack + next value

		case typeKeyReverseLookup:
			// peek first, if true continue
			if ptr < e.Top-2 && e.Code[ptr+2].T == typeKeyReverseLookupAttr {
				ptr++
				continue
			}
			result, err := store.KeyReverseLookup((*stack[top-1])[0])
			// store the addr of the result
			stack[top-1] = result
			// append the errors
			if err != nil {
				errs = append(errs, err)
			}

		case typeKeyReverseLookupAttr:
			// peek first, if true continue
			if ptr < e.Top-2 && e.Code[ptr+2].T == typeKeyReverseLookupHint {
				ptr++
				continue
			}
			result, err := store.KeyReverseLookupAttr((*stack[top-2])[0], (*stack[top-1])[0])
			// store the addr of the result
			stack[top-2] = result
			// reset top to nil, that value is no more useful to us
			stack[top-1] = nil
			// binary operator, pops of 2 elements to produce a result
			top--
			// append the errors
			if err != nil {
				errs = append(errs, err)
			}

		case typeKeyReverseLookupHint:
			result, err := store.KeyReverseLookupHint((*stack[top-3])[0], (*stack[top-2])[0], (*stack[top-1])[0])
			// store the addr of the result
			stack[top-3] = result
			// reset top to nil, that value is no more useful to us
			stack[top-2] = nil
			// reset top most to nil, that value is no more useful to us
			stack[top-1] = nil
			// tertiary operator, pops of 3 elements to produce a result
			top--
			top--
			// append the errors
			if err != nil {
				errs = append(errs, err)
			}

		// Range Set Operations
		// --------------------
		// All Set Operations are binary, pop off the stack the last two elements,
		// do the operation and push it back.
		case typeUnion:
			var result = make([]string, 0)
			rangeops.Union(stack[top-2], stack[top-1], &result)
			// store the addr of the result
			stack[top-2] = &result
			// reset top to nil, that value is no more useful to us
			stack[top-1] = nil
			top-- // merged two values to 1

		case typeIntersection:
			var result = make([]string, 0)
			rangeops.Intersection(stack[top-2], stack[top-1], &result)
			// store the addr of the result
			stack[top-2] = &result
			top-- // merged two values to 1

		case typeDifference:
			var result = make([]string, 0)
			rangeops.Difference(stack[top-2], stack[top-1], &result)
			// store the addr of the result
			stack[top-2] = &result
			top-- // merged two values to 1

		} // switch

		ptr++
	}

	return stack[0], errs
}
