package rangeexpr

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 4

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	rulee
	rulecombinedexpr
	rulerexpr
	rulecexpr
	ruleunion
	ruleintersect
	ruledifference
	rulecluster
	rulekey
	rulerlookup
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
	ruleAction3
	rulePegText
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"e",
	"combinedexpr",
	"rexpr",
	"cexpr",
	"union",
	"intersect",
	"difference",
	"cluster",
	"key",
	"rlookup",
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
	"Action3",
	"PegText",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",

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
	rules  [29]func() bool
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
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)

		case ruleAction0:
			p.addOperator(typeUnion)
		case ruleAction1:
			p.addOperator(typeIntersect)
		case ruleAction2:
			p.addOperator(typeDifference)
		case ruleAction3:
			p.addClusterOperator(typeClusterLookup)
		case ruleAction4:
			p.addClusterOperator(typeKeyLookup)
			p.addValue(buffer[begin:end])
		case ruleAction5:
			p.addClusterOperator(typeKeyReverseLookup)
		case ruleAction6:
			p.addClusterOperator(typeKeyReverseLookupAttr)
			p.addValue(buffer[begin:end])
		case ruleAction7:
			p.addClusterOperator(typeKeyReverseLookupHint)
		case ruleAction8:
			p.addValue(buffer[begin:end])

		}
	}
	_, _, _ = buffer, begin, end
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
		/* 1 combinedexpr <- <(rexpr cexpr?)> */
		func() bool {
			position5, tokenIndex5, depth5 := position, tokenIndex, depth
			{
				position6 := position
				depth++
				if !_rules[rulerexpr]() {
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
		/* 2 rexpr <- <(sp (cluster / brackets / value / rlookup))> */
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
						if buffer[position] != rune('%') {
							goto l12
						}
						position++
						{
							position14, tokenIndex14, depth14 := position, tokenIndex, depth
							if !_rules[rulerexpr]() {
								goto l14
							}
							goto l15
						l14:
							position, tokenIndex, depth = position14, tokenIndex14, depth14
						}
					l15:
						{
							add(ruleAction3, position)
						}
						{
							position17, tokenIndex17, depth17 := position, tokenIndex, depth
							{
								position19 := position
								depth++
								if buffer[position] != rune(':') {
									goto l17
								}
								position++
								{
									position20 := position
									depth++
									{
										position23, tokenIndex23, depth23 := position, tokenIndex, depth
										if c := buffer[position]; c < rune('A') || c > rune('Z') {
											goto l24
										}
										position++
										goto l23
									l24:
										position, tokenIndex, depth = position23, tokenIndex23, depth23
										if c := buffer[position]; c < rune('0') || c > rune('9') {
											goto l17
										}
										position++
									}
								l23:
								l21:
									{
										position22, tokenIndex22, depth22 := position, tokenIndex, depth
										{
											position25, tokenIndex25, depth25 := position, tokenIndex, depth
											if c := buffer[position]; c < rune('A') || c > rune('Z') {
												goto l26
											}
											position++
											goto l25
										l26:
											position, tokenIndex, depth = position25, tokenIndex25, depth25
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l22
											}
											position++
										}
									l25:
										goto l21
									l22:
										position, tokenIndex, depth = position22, tokenIndex22, depth22
									}
									depth--
									add(rulePegText, position20)
								}
								{
									add(ruleAction4, position)
								}
								depth--
								add(rulekey, position19)
							}
							goto l18
						l17:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
						}
					l18:
						depth--
						add(rulecluster, position13)
					}
					goto l11
				l12:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
					{
						position29 := position
						depth++
						if buffer[position] != rune('(') {
							goto l28
						}
						position++
						if !_rules[rulecombinedexpr]() {
							goto l28
						}
						if buffer[position] != rune(')') {
							goto l28
						}
						position++
						depth--
						add(rulebrackets, position29)
					}
					goto l11
				l28:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
					if !_rules[rulevalue]() {
						goto l30
					}
					goto l11
				l30:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
					{
						position31 := position
						depth++
						if buffer[position] != rune('*') {
							goto l9
						}
						position++
						if !_rules[rulevalue]() {
							goto l9
						}
						{
							add(ruleAction5, position)
						}
						{
							position33, tokenIndex33, depth33 := position, tokenIndex, depth
							{
								position35 := position
								depth++
								if buffer[position] != rune(';') {
									goto l33
								}
								position++
								{
									position36 := position
									depth++
									{
										position39, tokenIndex39, depth39 := position, tokenIndex, depth
										if c := buffer[position]; c < rune('A') || c > rune('Z') {
											goto l40
										}
										position++
										goto l39
									l40:
										position, tokenIndex, depth = position39, tokenIndex39, depth39
										if c := buffer[position]; c < rune('0') || c > rune('9') {
											goto l33
										}
										position++
									}
								l39:
								l37:
									{
										position38, tokenIndex38, depth38 := position, tokenIndex, depth
										{
											position41, tokenIndex41, depth41 := position, tokenIndex, depth
											if c := buffer[position]; c < rune('A') || c > rune('Z') {
												goto l42
											}
											position++
											goto l41
										l42:
											position, tokenIndex, depth = position41, tokenIndex41, depth41
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l38
											}
											position++
										}
									l41:
										goto l37
									l38:
										position, tokenIndex, depth = position38, tokenIndex38, depth38
									}
									depth--
									add(rulePegText, position36)
								}
								{
									add(ruleAction6, position)
								}
								{
									position44, tokenIndex44, depth44 := position, tokenIndex, depth
									{
										position46 := position
										depth++
										{
											add(ruleAction7, position)
										}
										if buffer[position] != rune(':') {
											goto l44
										}
										position++
										if !_rules[rulevalue]() {
											goto l44
										}
										depth--
										add(rulehint, position46)
									}
									goto l45
								l44:
									position, tokenIndex, depth = position44, tokenIndex44, depth44
								}
							l45:
								depth--
								add(ruleattr, position35)
							}
							goto l34
						l33:
							position, tokenIndex, depth = position33, tokenIndex33, depth33
						}
					l34:
						depth--
						add(rulerlookup, position31)
					}
				}
			l11:
				depth--
				add(rulerexpr, position10)
			}
			return true
		l9:
			position, tokenIndex, depth = position9, tokenIndex9, depth9
			return false
		},
		/* 3 cexpr <- <(sp (union / intersect / difference) sp)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if !_rules[rulesp]() {
					goto l48
				}
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					{
						position52 := position
						depth++
						if buffer[position] != rune(',') {
							goto l51
						}
						position++
						if !_rules[rulerexpr]() {
							goto l51
						}
						{
							position53, tokenIndex53, depth53 := position, tokenIndex, depth
							if !_rules[rulecexpr]() {
								goto l53
							}
							goto l54
						l53:
							position, tokenIndex, depth = position53, tokenIndex53, depth53
						}
					l54:
						{
							add(ruleAction0, position)
						}
						depth--
						add(ruleunion, position52)
					}
					goto l50
				l51:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					{
						position57 := position
						depth++
						if buffer[position] != rune('&') {
							goto l56
						}
						position++
						if !_rules[rulerexpr]() {
							goto l56
						}
						{
							position58, tokenIndex58, depth58 := position, tokenIndex, depth
							if !_rules[rulecexpr]() {
								goto l58
							}
							goto l59
						l58:
							position, tokenIndex, depth = position58, tokenIndex58, depth58
						}
					l59:
						{
							add(ruleAction1, position)
						}
						depth--
						add(ruleintersect, position57)
					}
					goto l50
				l56:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					{
						position61 := position
						depth++
						if buffer[position] != rune('^') {
							goto l48
						}
						position++
						if !_rules[rulerexpr]() {
							goto l48
						}
						{
							position62, tokenIndex62, depth62 := position, tokenIndex, depth
							if !_rules[rulecexpr]() {
								goto l62
							}
							goto l63
						l62:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
						}
					l63:
						{
							add(ruleAction2, position)
						}
						depth--
						add(ruledifference, position61)
					}
				}
			l50:
				if !_rules[rulesp]() {
					goto l48
				}
				depth--
				add(rulecexpr, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 4 union <- <(',' rexpr cexpr? Action0)> */
		nil,
		/* 5 intersect <- <('&' rexpr cexpr? Action1)> */
		nil,
		/* 6 difference <- <('^' rexpr cexpr? Action2)> */
		nil,
		/* 7 cluster <- <('%' rexpr? Action3 key?)> */
		nil,
		/* 8 key <- <(':' <([A-Z] / [0-9])+> Action4)> */
		nil,
		/* 9 rlookup <- <('*' value Action5 attr?)> */
		nil,
		/* 10 attr <- <(';' <([A-Z] / [0-9])+> Action6 hint?)> */
		nil,
		/* 11 hint <- <(Action7 ':' value)> */
		nil,
		/* 12 value <- <(<((first middle+) / (first last?))> Action8)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				{
					position75 := position
					depth++
					{
						position76, tokenIndex76, depth76 := position, tokenIndex, depth
						if !_rules[rulefirst]() {
							goto l77
						}
						{
							position80 := position
							depth++
							if buffer[position] != rune('-') {
								goto l77
							}
							position++
							if !_rules[rulelast]() {
								goto l77
							}
							depth--
							add(rulemiddle, position80)
						}
					l78:
						{
							position79, tokenIndex79, depth79 := position, tokenIndex, depth
							{
								position81 := position
								depth++
								if buffer[position] != rune('-') {
									goto l79
								}
								position++
								if !_rules[rulelast]() {
									goto l79
								}
								depth--
								add(rulemiddle, position81)
							}
							goto l78
						l79:
							position, tokenIndex, depth = position79, tokenIndex79, depth79
						}
						goto l76
					l77:
						position, tokenIndex, depth = position76, tokenIndex76, depth76
						if !_rules[rulefirst]() {
							goto l73
						}
						{
							position82, tokenIndex82, depth82 := position, tokenIndex, depth
							if !_rules[rulelast]() {
								goto l82
							}
							goto l83
						l82:
							position, tokenIndex, depth = position82, tokenIndex82, depth82
						}
					l83:
					}
				l76:
					depth--
					add(rulePegText, position75)
				}
				{
					add(ruleAction8, position)
				}
				depth--
				add(rulevalue, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 13 first <- <[a-z]+> */
		func() bool {
			position85, tokenIndex85, depth85 := position, tokenIndex, depth
			{
				position86 := position
				depth++
				if c := buffer[position]; c < rune('a') || c > rune('z') {
					goto l85
				}
				position++
			l87:
				{
					position88, tokenIndex88, depth88 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l88
					}
					position++
					goto l87
				l88:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
				}
				depth--
				add(rulefirst, position86)
			}
			return true
		l85:
			position, tokenIndex, depth = position85, tokenIndex85, depth85
			return false
		},
		/* 14 middle <- <('-' last)> */
		nil,
		/* 15 last <- <([a-z] / [0-9])*> */
		func() bool {
			{
				position91 := position
				depth++
			l92:
				{
					position93, tokenIndex93, depth93 := position, tokenIndex, depth
					{
						position94, tokenIndex94, depth94 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l95
						}
						position++
						goto l94
					l95:
						position, tokenIndex, depth = position94, tokenIndex94, depth94
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l93
						}
						position++
					}
				l94:
					goto l92
				l93:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
				}
				depth--
				add(rulelast, position91)
			}
			return true
		},
		/* 16 brackets <- <('(' combinedexpr ')')> */
		nil,
		/* 17 sp <- <' '*> */
		func() bool {
			{
				position98 := position
				depth++
			l99:
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l100
					}
					position++
					goto l99
				l100:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
				}
				depth--
				add(rulesp, position98)
			}
			return true
		},
		/* 19 Action0 <- <{ p.addOperator(typeUnion) }> */
		nil,
		/* 20 Action1 <- <{ p.addOperator(typeIntersect) }> */
		nil,
		/* 21 Action2 <- <{ p.addOperator(typeDifference) }> */
		nil,
		/* 22 Action3 <- <{ p.addClusterOperator(typeClusterLookup) }> */
		nil,
		nil,
		/* 24 Action4 <- <{ p.addClusterOperator(typeKeyLookup); p.addValue(buffer[begin:end]); }> */
		nil,
		/* 25 Action5 <- <{ p.addClusterOperator(typeKeyReverseLookup) }> */
		nil,
		/* 26 Action6 <- <{ p.addClusterOperator(typeKeyReverseLookupAttr); p.addValue(buffer[begin:end]); }> */
		nil,
		/* 27 Action7 <- <{ p.addClusterOperator(typeKeyReverseLookupHint) }> */
		nil,
		/* 28 Action8 <- <{ p.addValue(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
