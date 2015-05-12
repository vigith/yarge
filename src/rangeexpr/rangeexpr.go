package rangeexpr

import (
	"fmt"
)

type Type uint8

const (
	typeNumber Type = iota
	typeUnion
	typeIntersect
	typeDifference
	typeClusterLookup
	typeKeyLookup
	typeKeyReverseLookup
	typeKeyReverseLookupAttr
	typeKeyReverseLookupHint
)

type ByteCode struct {
	T     Type
	Value string
}

type Expression struct {
	Code []ByteCode
	Top  int
}

func (e *Expression) Init(expression string) {
	e.Code = make([]ByteCode, len(expression))
}

func (e *Expression) addClusterOperator(operator Type) {
	code, top := e.Code, e.Top
	e.Top++
	code[top].T = operator
}

func (e *Expression) addOperator(operator Type) {
	code, top := e.Code, e.Top
	e.Top++
	code[top].T = operator
}

func (e *Expression) addValue(value string) {
	code, top := e.Code, e.Top
	e.Top++
	code[top].Value = value
}

func (e *Expression) Evaluate() *string {
	stack, top := make([]string, len(e.Code)), 0
	for _, code := range e.Code[0:e.Top] {
		fmt.Println(code, stack, top)
		top--
	}
	//return &stack[0]
	return nil
}
