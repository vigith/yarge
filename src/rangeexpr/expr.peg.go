package rangeexpr

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	rulee
	rulecombinedexpr
	ruleyrexpr
	rulecexpr
	ruleunion
	ruleintersection
	ruledifference
	rulecluster
	rulekey
	rulerlookup
	rulervalue
	ruleattr
	rulehint
	rulevalue
	rulefirst
	rulemiddle
	rulelast
	rulebrackets
	rulesp
	ruleAction0
	ruleAction1
	ruleAction2
	rulePegText
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"e",
	"combinedexpr",
	"yrexpr",
	"cexpr",
	"union",
	"intersection",
	"difference",
	"cluster",
	"key",
	"rlookup",
	"rvalue",
	"attr",
	"hint",
	"value",
	"first",
	"middle",
	"last",
	"brackets",
	"sp",
	"Action0",
	"Action1",
	"Action2",
	"PegText",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next uint32, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/*func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2 * len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}*/

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type RangeExpr struct {
	Expression

	Buffer string
	buffer []rune
	rules  [32]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *RangeExpr
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *RangeExpr) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *RangeExpr) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *RangeExpr) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.addOperator(typeUnion)
		case ruleAction1:
			p.addOperator(typeIntersection)
		case ruleAction2:
			p.addOperator(typeDifference)
		case ruleAction3:
			p.addValue(buffer[begin:end])
		case ruleAction4:
			p.addOperator(typeClusterLookup)
		case ruleAction5:
			p.addValue(buffer[begin:end])
			p.addOperator(typeKeyLookup)
		case ruleAction6:
			p.addOperator(typeKeyReverseLookup)
		case ruleAction7:
			p.addValue(buffer[begin:end])
		case ruleAction8:
			p.addValue(buffer[begin:end])
			p.addOperator(typeKeyReverseLookupAttr)
		case ruleAction9:
			p.addOperator(typeKeyReverseLookupHint)
		case ruleAction10:
			p.addValue(buffer[begin:end])

		}
	}
	_, _, _, _ = buffer, text, begin, end
}

func (p *RangeExpr) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 e <- <(combinedexpr? !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !_rules[rulecombinedexpr]() {
						goto l2
					}
					goto l3
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
			l3:
				{
					position4, tokenIndex4, depth4 := position, tokenIndex, depth
					if !matchDot() {
						goto l4
					}
					goto l0
				l4:
					position, tokenIndex, depth = position4, tokenIndex4, depth4
				}
				depth--
				add(rulee, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 combinedexpr <- <(yrexpr cexpr?)> */
		func() bool {
			position5, tokenIndex5, depth5 := position, tokenIndex, depth
			{
				position6 := position
				depth++
				if !_rules[ruleyrexpr]() {
					goto l5
				}
				{
					position7, tokenIndex7, depth7 := position, tokenIndex, depth
					if !_rules[rulecexpr]() {
						goto l7
					}
					goto l8
				l7:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
				}
			l8:
				depth--
				add(rulecombinedexpr, position6)
			}
			return true
		l5:
			position, tokenIndex, depth = position5, tokenIndex5, depth5
			return false
		},
		/* 2 yrexpr <- <(sp (brackets / cluster / value / rlookup))> */
		func() bool {
			position9, tokenIndex9, depth9 := position, tokenIndex, depth
			{
				position10 := position
				depth++
				if !_rules[rulesp]() {
					goto l9
				}
				{
					position11, tokenIndex11, depth11 := position, tokenIndex, depth
					{
						position13 := position
						depth++
						if buffer[position] != rune('(') {
							goto l12
						}
						position++
						if !_rules[rulecombinedexpr]() {
							goto l12
						}
						if buffer[position] != rune(')') {
							goto l12
						}
						position++
						depth--
						add(rulebrackets, position13)
					}
					goto l11
				l12:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
					{
						position15 := position
						depth++
						{
							position16, tokenIndex16, depth16 := position, tokenIndex, depth
							if buffer[position] != rune('%') {
								goto l17
							}
							position++
							{
								position18 := position
								depth++
								if buffer[position] != rune('R') {
									goto l17
								}
								position++
								if buffer[position] != rune('A') {
									goto l17
								}
								position++
								if buffer[position] != rune('N') {
									goto l17
								}
								position++
								if buffer[position] != rune('G') {
									goto l17
								}
								position++
								if buffer[position] != rune('E') {
									goto l17
								}
								position++
								depth--
								add(rulePegText, position18)
							}
							{
								add(ruleAction3, position)
							}
							goto l16
						l17:
							position, tokenIndex, depth = position16, tokenIndex16, depth16
							if buffer[position] != rune('%') {
								goto l20
							}
							position++
							if !_rules[ruleyrexpr]() {
								goto l20
							}
							goto l16
						l20:
							position, tokenIndex, depth = position16, tokenIndex16, depth16
							if buffer[position] != rune('%') {
								goto l14
							}
							position++
							if !_rules[rulerlookup]() {
								goto l14
							}
						}
					l16:
						{
							add(ruleAction4, position)
						}
						{
							position22, tokenIndex22, depth22 := position, tokenIndex, depth
							{
								position24 := position
								depth++
								if buffer[position] != rune(':') {
									goto l22
								}
								position++
								{
									position25 := position
									depth++
									{
										position28, tokenIndex28, depth28 := position, tokenIndex, depth
										if c := buffer[position]; c < rune('A') || c > rune('Z') {
											goto l29
										}
										position++
										goto l28
									l29:
										position, tokenIndex, depth = position28, tokenIndex28, depth28
										if c := buffer[position]; c < rune('0') || c > rune('9') {
											goto l22
										}
										position++
									}
								l28:
								l26:
									{
										position27, tokenIndex27, depth27 := position, tokenIndex, depth
										{
											position30, tokenIndex30, depth30 := position, tokenIndex, depth
											if c := buffer[position]; c < rune('A') || c > rune('Z') {
												goto l31
											}
											position++
											goto l30
										l31:
											position, tokenIndex, depth = position30, tokenIndex30, depth30
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l27
											}
											position++
										}
									l30:
										goto l26
									l27:
										position, tokenIndex, depth = position27, tokenIndex27, depth27
									}
									depth--
									add(rulePegText, position25)
								}
								{
									add(ruleAction5, position)
								}
								depth--
								add(rulekey, position24)
							}
							goto l23
						l22:
							position, tokenIndex, depth = position22, tokenIndex22, depth22
						}
					l23:
						depth--
						add(rulecluster, position15)
					}
					goto l11
				l14:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
					if !_rules[rulevalue]() {
						goto l33
					}
					goto l11
				l33:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
					if !_rules[rulerlookup]() {
						goto l9
					}
				}
			l11:
				depth--
				add(ruleyrexpr, position10)
			}
			return true
		l9:
			position, tokenIndex, depth = position9, tokenIndex9, depth9
			return false
		},
		/* 3 cexpr <- <(sp (union / intersection / difference) sp)> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				if !_rules[rulesp]() {
					goto l34
				}
				{
					position36, tokenIndex36, depth36 := position, tokenIndex, depth
					{
						position38 := position
						depth++
						if buffer[position] != rune(',') {
							goto l37
						}
						position++
						if !_rules[ruleyrexpr]() {
							goto l37
						}
						{
							position39, tokenIndex39, depth39 := position, tokenIndex, depth
							if !_rules[rulecexpr]() {
								goto l39
							}
							goto l40
						l39:
							position, tokenIndex, depth = position39, tokenIndex39, depth39
						}
					l40:
						{
							add(ruleAction0, position)
						}
						depth--
						add(ruleunion, position38)
					}
					goto l36
				l37:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
					{
						position43 := position
						depth++
						if buffer[position] != rune(',') {
							goto l42
						}
						position++
						if !_rules[rulesp]() {
							goto l42
						}
						if buffer[position] != rune('&') {
							goto l42
						}
						position++
						if !_rules[ruleyrexpr]() {
							goto l42
						}
						{
							position44, tokenIndex44, depth44 := position, tokenIndex, depth
							if !_rules[rulecexpr]() {
								goto l44
							}
							goto l45
						l44:
							position, tokenIndex, depth = position44, tokenIndex44, depth44
						}
					l45:
						{
							add(ruleAction1, position)
						}
						depth--
						add(ruleintersection, position43)
					}
					goto l36
				l42:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
					{
						position47 := position
						depth++
						if buffer[position] != rune(',') {
							goto l34
						}
						position++
						if !_rules[rulesp]() {
							goto l34
						}
						if buffer[position] != rune('-') {
							goto l34
						}
						position++
						if !_rules[ruleyrexpr]() {
							goto l34
						}
						{
							position48, tokenIndex48, depth48 := position, tokenIndex, depth
							if !_rules[rulecexpr]() {
								goto l48
							}
							goto l49
						l48:
							position, tokenIndex, depth = position48, tokenIndex48, depth48
						}
					l49:
						{
							add(ruleAction2, position)
						}
						depth--
						add(ruledifference, position47)
					}
				}
			l36:
				if !_rules[rulesp]() {
					goto l34
				}
				depth--
				add(rulecexpr, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 4 union <- <(',' yrexpr cexpr? Action0)> */
		nil,
		/* 5 intersection <- <(',' sp '&' yrexpr cexpr? Action1)> */
		nil,
		/* 6 difference <- <(',' sp '-' yrexpr cexpr? Action2)> */
		nil,
		/* 7 cluster <- <((('%' <('R' 'A' 'N' 'G' 'E')> Action3) / ('%' yrexpr) / ('%' rlookup)) Action4 key?)> */
		nil,
		/* 8 key <- <(':' <([A-Z] / [0-9])+> Action5)> */
		nil,
		/* 9 rlookup <- <('*' rvalue Action6 attr? cexpr?)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				if buffer[position] != rune('*') {
					goto l56
				}
				position++
				{
					position58 := position
					depth++
					{
						position59 := position
						depth++
						{
							position62, tokenIndex62, depth62 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l63
							}
							position++
							goto l62
						l63:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l64
							}
							position++
							goto l62
						l64:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							{
								position66, tokenIndex66, depth66 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l67
								}
								position++
								goto l66
							l67:
								position, tokenIndex, depth = position66, tokenIndex66, depth66
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l65
								}
								position++
							}
						l66:
							goto l62
						l65:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if buffer[position] != rune('-') {
								goto l68
							}
							position++
							goto l62
						l68:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if buffer[position] != rune(' ') {
								goto l69
							}
							position++
							goto l62
						l69:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							if buffer[position] != rune('.') {
								goto l56
							}
							position++
						}
					l62:
					l60:
						{
							position61, tokenIndex61, depth61 := position, tokenIndex, depth
							{
								position70, tokenIndex70, depth70 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l71
								}
								position++
								goto l70
							l71:
								position, tokenIndex, depth = position70, tokenIndex70, depth70
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l72
								}
								position++
								goto l70
							l72:
								position, tokenIndex, depth = position70, tokenIndex70, depth70
								{
									position74, tokenIndex74, depth74 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l75
									}
									position++
									goto l74
								l75:
									position, tokenIndex, depth = position74, tokenIndex74, depth74
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l73
									}
									position++
								}
							l74:
								goto l70
							l73:
								position, tokenIndex, depth = position70, tokenIndex70, depth70
								if buffer[position] != rune('-') {
									goto l76
								}
								position++
								goto l70
							l76:
								position, tokenIndex, depth = position70, tokenIndex70, depth70
								if buffer[position] != rune(' ') {
									goto l77
								}
								position++
								goto l70
							l77:
								position, tokenIndex, depth = position70, tokenIndex70, depth70
								if buffer[position] != rune('.') {
									goto l61
								}
								position++
							}
						l70:
							goto l60
						l61:
							position, tokenIndex, depth = position61, tokenIndex61, depth61
						}
						depth--
						add(rulePegText, position59)
					}
					{
						add(ruleAction7, position)
					}
					depth--
					add(rulervalue, position58)
				}
				{
					add(ruleAction6, position)
				}
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					{
						position82 := position
						depth++
						if buffer[position] != rune(';') {
							goto l80
						}
						position++
						{
							position83 := position
							depth++
							{
								position86, tokenIndex86, depth86 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l87
								}
								position++
								goto l86
							l87:
								position, tokenIndex, depth = position86, tokenIndex86, depth86
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l80
								}
								position++
							}
						l86:
						l84:
							{
								position85, tokenIndex85, depth85 := position, tokenIndex, depth
								{
									position88, tokenIndex88, depth88 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('A') || c > rune('Z') {
										goto l89
									}
									position++
									goto l88
								l89:
									position, tokenIndex, depth = position88, tokenIndex88, depth88
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l85
									}
									position++
								}
							l88:
								goto l84
							l85:
								position, tokenIndex, depth = position85, tokenIndex85, depth85
							}
							depth--
							add(rulePegText, position83)
						}
						{
							add(ruleAction8, position)
						}
						{
							position91, tokenIndex91, depth91 := position, tokenIndex, depth
							{
								position93 := position
								depth++
								if buffer[position] != rune(':') {
									goto l91
								}
								position++
								if !_rules[rulevalue]() {
									goto l91
								}
								{
									add(ruleAction9, position)
								}
								depth--
								add(rulehint, position93)
							}
							goto l92
						l91:
							position, tokenIndex, depth = position91, tokenIndex91, depth91
						}
					l92:
						depth--
						add(ruleattr, position82)
					}
					goto l81
				l80:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
				}
			l81:
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					if !_rules[rulecexpr]() {
						goto l95
					}
					goto l96
				l95:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
				}
			l96:
				depth--
				add(rulerlookup, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 10 rvalue <- <(<([a-z] / [A-Z] / ([0-9] / [0-9]) / '-' / ' ' / '.')+> Action7)> */
		nil,
		/* 11 attr <- <(';' <([A-Z] / [0-9])+> Action8 hint?)> */
		nil,
		/* 12 hint <- <(':' value Action9)> */
		nil,
		/* 13 value <- <(<((first last? middle+) / (first last*))> Action10)> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				{
					position102 := position
					depth++
					{
						position103, tokenIndex103, depth103 := position, tokenIndex, depth
						if !_rules[rulefirst]() {
							goto l104
						}
						{
							position105, tokenIndex105, depth105 := position, tokenIndex, depth
							if !_rules[rulelast]() {
								goto l105
							}
							goto l106
						l105:
							position, tokenIndex, depth = position105, tokenIndex105, depth105
						}
					l106:
						{
							position109 := position
							depth++
							{
								position110, tokenIndex110, depth110 := position, tokenIndex, depth
								if buffer[position] != rune('-') {
									goto l111
								}
								position++
								goto l110
							l111:
								position, tokenIndex, depth = position110, tokenIndex110, depth110
								if buffer[position] != rune('.') {
									goto l104
								}
								position++
							}
						l110:
							if !_rules[rulelast]() {
								goto l104
							}
							depth--
							add(rulemiddle, position109)
						}
					l107:
						{
							position108, tokenIndex108, depth108 := position, tokenIndex, depth
							{
								position112 := position
								depth++
								{
									position113, tokenIndex113, depth113 := position, tokenIndex, depth
									if buffer[position] != rune('-') {
										goto l114
									}
									position++
									goto l113
								l114:
									position, tokenIndex, depth = position113, tokenIndex113, depth113
									if buffer[position] != rune('.') {
										goto l108
									}
									position++
								}
							l113:
								if !_rules[rulelast]() {
									goto l108
								}
								depth--
								add(rulemiddle, position112)
							}
							goto l107
						l108:
							position, tokenIndex, depth = position108, tokenIndex108, depth108
						}
						goto l103
					l104:
						position, tokenIndex, depth = position103, tokenIndex103, depth103
						if !_rules[rulefirst]() {
							goto l100
						}
					l115:
						{
							position116, tokenIndex116, depth116 := position, tokenIndex, depth
							if !_rules[rulelast]() {
								goto l116
							}
							goto l115
						l116:
							position, tokenIndex, depth = position116, tokenIndex116, depth116
						}
					}
				l103:
					depth--
					add(rulePegText, position102)
				}
				{
					add(ruleAction10, position)
				}
				depth--
				add(rulevalue, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 14 first <- <([a-z] / [0-9])+> */
		func() bool {
			position118, tokenIndex118, depth118 := position, tokenIndex, depth
			{
				position119 := position
				depth++
				{
					position122, tokenIndex122, depth122 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l123
					}
					position++
					goto l122
				l123:
					position, tokenIndex, depth = position122, tokenIndex122, depth122
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l118
					}
					position++
				}
			l122:
			l120:
				{
					position121, tokenIndex121, depth121 := position, tokenIndex, depth
					{
						position124, tokenIndex124, depth124 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l125
						}
						position++
						goto l124
					l125:
						position, tokenIndex, depth = position124, tokenIndex124, depth124
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l121
						}
						position++
					}
				l124:
					goto l120
				l121:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
				}
				depth--
				add(rulefirst, position119)
			}
			return true
		l118:
			position, tokenIndex, depth = position118, tokenIndex118, depth118
			return false
		},
		/* 15 middle <- <(('-' / '.') last)> */
		nil,
		/* 16 last <- <first> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				if !_rules[rulefirst]() {
					goto l127
				}
				depth--
				add(rulelast, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 17 brackets <- <('(' combinedexpr ')')> */
		nil,
		/* 18 sp <- <' '*> */
		func() bool {
			{
				position131 := position
				depth++
			l132:
				{
					position133, tokenIndex133, depth133 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l133
					}
					position++
					goto l132
				l133:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
				}
				depth--
				add(rulesp, position131)
			}
			return true
		},
		/* 20 Action0 <- <{ p.addOperator(typeUnion) }> */
		nil,
		/* 21 Action1 <- <{ p.addOperator(typeIntersection) }> */
		nil,
		/* 22 Action2 <- <{ p.addOperator(typeDifference) }> */
		nil,
		nil,
		/* 24 Action3 <- <{ p.addValue(buffer[begin:end]); }> */
		nil,
		/* 25 Action4 <- <{ p.addOperator(typeClusterLookup) }> */
		nil,
		/* 26 Action5 <- <{ p.addValue(buffer[begin:end]); p.addOperator(typeKeyLookup); }> */
		nil,
		/* 27 Action6 <- <{ p.addOperator(typeKeyReverseLookup); }> */
		nil,
		/* 28 Action7 <- <{ p.addValue(buffer[begin:end]) }> */
		nil,
		/* 29 Action8 <- <{ p.addValue(buffer[begin:end]); p.addOperator(typeKeyReverseLookupAttr); }> */
		nil,
		/* 30 Action9 <- <{ p.addOperator(typeKeyReverseLookupHint) }> */
		nil,
		/* 31 Action10 <- <{ p.addValue(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
